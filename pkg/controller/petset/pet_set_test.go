/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package petset

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/client/cache"
	fake_internal "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/apps/unversioned"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/apps/unversioned/fake"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/util/errors"
)

func newFakePetSetController() (*PetSetController, *fakePetClient) {
	fpc := newFakePetClient()
	return &PetSetController{
		kubeClient:       nil,
		blockingPetStore: newUnHealthyPetTracker(fpc),
		podStoreSynced:   func() bool { return true },
		psStore:          cache.StoreToPetSetLister{Store: cache.NewStore(controller.KeyFunc)},
		podStore:         cache.StoreToPodLister{Indexer: cache.NewIndexer(controller.KeyFunc, cache.Indexers{})},
		newSyncer: func(blockingPet *pcb) *petSyncer {
			return &petSyncer{fpc, blockingPet}
		},
	}, fpc
}

func checkPets(ps *apps.PetSet, creates, deletes int, fc *fakePetClient, t *testing.T) {
	if fc.petsCreated != creates || fc.petsDeleted != deletes {
		t.Errorf("Found (creates: %d, deletes: %d), expected (creates: %d, deletes: %d)", fc.petsCreated, fc.petsDeleted, creates, deletes)
	}
	gotClaims := map[string]api.PersistentVolumeClaim{}
	for _, pvc := range fc.claims {
		gotClaims[pvc.Name] = pvc
	}
	for i := range fc.pets {
		expectedPet, _ := newPCB(fmt.Sprintf("%v", i), ps)
		if identityHash(ps, fc.pets[i].pod) != identityHash(ps, expectedPet.pod) {
			t.Errorf("Unexpected pet at index %d", i)
		}
		for _, pvc := range expectedPet.pvcs {
			gotPVC, ok := gotClaims[pvc.Name]
			if !ok {
				t.Errorf("PVC %v not created for pet %v", pvc.Name, expectedPet.pod.Name)
			}
			if !reflect.DeepEqual(gotPVC.Spec, pvc.Spec) {
				t.Errorf("got PVC %v differs from created pvc", pvc.Name)
			}
		}
	}
}

func scalePetSet(t *testing.T, ps *apps.PetSet, psc *PetSetController, fc *fakePetClient, scale int) error {
	errs := []error{}
	for i := 0; i < scale; i++ {
		pl := fc.getPodList()
		if len(pl) != i {
			t.Errorf("Unexpected number of pets, expected %d found %d", i, len(pl))
		}
		if _, syncErr := psc.syncPetSet(ps, pl); syncErr != nil {
			errs = append(errs, syncErr)
		}
		fc.setHealthy(i)
		checkPets(ps, i+1, 0, fc, t)
	}
	return errors.NewAggregate(errs)
}

func saturatePetSet(t *testing.T, ps *apps.PetSet, psc *PetSetController, fc *fakePetClient) {
	err := scalePetSet(t, ps, psc, fc, ps.Spec.Replicas)
	if err != nil {
		t.Errorf("Error scalePetSet: %v", err)
	}
}

func TestPetSetControllerCreates(t *testing.T) {
	psc, fc := newFakePetSetController()
	replicas := 3
	ps := newPetSet(replicas)

	saturatePetSet(t, ps, psc, fc)

	podList := fc.getPodList()
	// Deleted pet gets recreated
	fc.pets = fc.pets[:replicas-1]
	if _, err := psc.syncPetSet(ps, podList); err != nil {
		t.Errorf("Error syncing PetSet: %v", err)
	}
	checkPets(ps, replicas+1, 0, fc, t)
}

func TestPetSetControllerDeletes(t *testing.T) {
	psc, fc := newFakePetSetController()
	replicas := 4
	ps := newPetSet(replicas)

	saturatePetSet(t, ps, psc, fc)

	// Drain
	errs := []error{}
	ps.Spec.Replicas = 0
	knownPods := fc.getPodList()
	for i := replicas - 1; i >= 0; i-- {
		if len(fc.pets) != i+1 {
			t.Errorf("Unexpected number of pets, expected %d found %d", i+1, len(fc.pets))
		}
		if _, syncErr := psc.syncPetSet(ps, knownPods); syncErr != nil {
			errs = append(errs, syncErr)
		}
	}
	if len(errs) != 0 {
		t.Errorf("Error syncing PetSet: %v", errors.NewAggregate(errs))
	}
	checkPets(ps, replicas, replicas, fc, t)
}

func TestPetSetControllerRespectsTermination(t *testing.T) {
	psc, fc := newFakePetSetController()
	replicas := 4
	ps := newPetSet(replicas)

	saturatePetSet(t, ps, psc, fc)

	fc.setDeletionTimestamp(replicas - 1)
	ps.Spec.Replicas = 2
	_, err := psc.syncPetSet(ps, fc.getPodList())
	if err != nil {
		t.Errorf("Error syncing PetSet: %v", err)
	}
	// Finding a pod with the deletion timestamp will pause all deletions.
	knownPods := fc.getPodList()
	if len(knownPods) != 4 {
		t.Errorf("Pods deleted prematurely before deletion timestamp expired, len %d", len(knownPods))
	}
	fc.pets = fc.pets[:replicas-1]
	_, err = psc.syncPetSet(ps, fc.getPodList())
	if err != nil {
		t.Errorf("Error syncing PetSet: %v", err)
	}
	checkPets(ps, replicas, 1, fc, t)
}

func TestPetSetControllerRespectsOrder(t *testing.T) {
	psc, fc := newFakePetSetController()
	replicas := 4
	ps := newPetSet(replicas)

	saturatePetSet(t, ps, psc, fc)

	errs := []error{}
	ps.Spec.Replicas = 0
	// Shuffle known list and check that pets are deleted in reverse
	knownPods := fc.getPodList()
	for i := range knownPods {
		j := rand.Intn(i + 1)
		knownPods[i], knownPods[j] = knownPods[j], knownPods[i]
	}

	for i := 0; i < replicas; i++ {
		if len(fc.pets) != replicas-i {
			t.Errorf("Unexpected number of pets, expected %d found %d", i, len(fc.pets))
		}
		if _, syncErr := psc.syncPetSet(ps, knownPods); syncErr != nil {
			errs = append(errs, syncErr)
		}
		checkPets(ps, replicas, i+1, fc, t)
	}
	if len(errs) != 0 {
		t.Errorf("Error syncing PetSet: %v", errors.NewAggregate(errs))
	}
}

func TestPetSetControllerBlocksScaling(t *testing.T) {
	psc, fc := newFakePetSetController()
	replicas := 5
	ps := newPetSet(replicas)
	scalePetSet(t, ps, psc, fc, 3)

	// Create 4th pet, then before flipping it to healthy, kill the first pet.
	// There should only be 1 not-healty pet at a time.
	pl := fc.getPodList()
	if _, err := psc.syncPetSet(ps, pl); err != nil {
		t.Errorf("Error syncing PetSet: %v", err)
	}

	deletedPod := pl[0]
	fc.deletePetAtIndex(0)
	pl = fc.getPodList()
	if _, err := psc.syncPetSet(ps, pl); err != nil {
		t.Errorf("Error syncing PetSet: %v", err)
	}
	newPodList := fc.getPodList()
	for _, p := range newPodList {
		if p.Name == deletedPod.Name {
			t.Errorf("Deleted pod was created while existing pod was unhealthy")
		}
	}

	fc.setHealthy(len(newPodList) - 1)
	if _, err := psc.syncPetSet(ps, pl); err != nil {
		t.Errorf("Error syncing PetSet: %v", err)
	}

	found := false
	for _, p := range fc.getPodList() {
		if p.Name == deletedPod.Name {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Deleted pod was not created after existing pods became healthy")
	}
}

func TestPetSetBlockingPetIsCleared(t *testing.T) {
	psc, fc := newFakePetSetController()
	ps := newPetSet(3)
	scalePetSet(t, ps, psc, fc, 1)

	if blocking, err := psc.blockingPetStore.Get(ps, fc.getPodList()); err != nil || blocking != nil {
		t.Errorf("Unexpected blocking pet %v, err %v", blocking, err)
	}

	// 1 not yet healthy pet
	psc.syncPetSet(ps, fc.getPodList())

	if blocking, err := psc.blockingPetStore.Get(ps, fc.getPodList()); err != nil || blocking == nil {
		t.Errorf("Expected blocking pet %v, err %v", blocking, err)
	}

	// Deleting the petset should clear the blocking pet
	if err := psc.psStore.Store.Delete(ps); err != nil {
		t.Fatalf("Unable to delete pet %v from petset controller store.", ps.Name)
	}
	if err := psc.Sync(fmt.Sprintf("%v/%v", ps.Namespace, ps.Name)); err != nil {
		t.Errorf("Error during sync of deleted petset %v", err)
	}
	fc.pets = []*pcb{}
	fc.petsCreated = 0
	if blocking, err := psc.blockingPetStore.Get(ps, fc.getPodList()); err != nil || blocking != nil {
		t.Errorf("Unexpected blocking pet %v, err %v", blocking, err)
	}
	saturatePetSet(t, ps, psc, fc)

	// Make sure we don't leak the final blockin pet in the store
	psc.syncPetSet(ps, fc.getPodList())
	if p, exists, err := psc.blockingPetStore.store.GetByKey(fmt.Sprintf("%v/%v", ps.Namespace, ps.Name)); err != nil || exists {
		t.Errorf("Unexpected blocking pet, err %v: %+v", err, p)
	}
}

func TestSyncPetSetBlockedPet(t *testing.T) {
	psc, fc := newFakePetSetController()
	ps := newPetSet(3)
	i, _ := psc.syncPetSet(ps, fc.getPodList())
	if i != len(fc.getPodList()) {
		t.Errorf("syncPetSet should return actual amount of pods")
	}
}

type fakeClient struct {
	fake_internal.Clientset
	petSetClient *fakePetSetClient
}

func (c *fakeClient) Apps() unversioned.AppsInterface {
	return &fakeApps{c, &fake.FakeApps{}}
}

type fakeApps struct {
	*fakeClient
	*fake.FakeApps
}

func (c *fakeApps) PetSets(namespace string) unversioned.PetSetInterface {
	c.petSetClient.Namespace = namespace
	return c.petSetClient
}

type fakePetSetClient struct {
	*fake.FakePetSets
	Namespace string
	replicas  int
}

func (f *fakePetSetClient) UpdateStatus(petSet *apps.PetSet) (*apps.PetSet, error) {
	f.replicas = petSet.Status.Replicas
	return petSet, nil
}

func TestPetSetReplicaCount(t *testing.T) {
	fpsc := &fakePetSetClient{}
	psc, _ := newFakePetSetController()
	psc.kubeClient = &fakeClient{
		petSetClient: fpsc,
	}

	ps := newPetSet(3)
	psKey := fmt.Sprintf("%v/%v", ps.Namespace, ps.Name)
	psc.psStore.Store.Add(ps)

	if err := psc.Sync(psKey); err != nil {
		t.Errorf("Error during sync of deleted petset %v", err)
	}

	if fpsc.replicas != 1 {
		t.Errorf("Replicas count sent as status update for PetSet should be 1, is %d instead", fpsc.replicas)
	}
}
