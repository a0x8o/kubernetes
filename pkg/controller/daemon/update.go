/*
Copyright 2017 The Kubernetes Authors.

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

package daemon

import (
	"fmt"
	"sort"

	"github.com/golang/glog"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	apps "k8s.io/kubernetes/pkg/apis/apps/v1beta1"
	extensions "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/daemon/util"
	labelsutil "k8s.io/kubernetes/pkg/util/labels"
)

// rollingUpdate deletes old daemon set pods making sure that no more than
// ds.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable pods are unavailable
func (dsc *DaemonSetsController) rollingUpdate(ds *extensions.DaemonSet, hash string) error {
	nodeToDaemonPods, err := dsc.getNodesToDaemonPods(ds)
	if err != nil {
		return fmt.Errorf("couldn't get node to daemon pod mapping for daemon set %q: %v", ds.Name, err)
	}

	_, oldPods := dsc.getAllDaemonSetPods(ds, nodeToDaemonPods, hash)
	maxUnavailable, numUnavailable, err := dsc.getUnavailableNumbers(ds, nodeToDaemonPods)
	if err != nil {
		return fmt.Errorf("Couldn't get unavailable numbers: %v", err)
	}
	oldAvailablePods, oldUnavailablePods := util.SplitByAvailablePods(ds.Spec.MinReadySeconds, oldPods)

	// for oldPods delete all not running pods
	var oldPodsToDelete []string
	glog.V(4).Infof("Marking all unavailable old pods for deletion")
	for _, pod := range oldUnavailablePods {
		// Skip terminating pods. We won't delete them again
		if pod.DeletionTimestamp != nil {
			continue
		}
		glog.V(4).Infof("Marking pod %s/%s for deletion", ds.Name, pod.Name)
		oldPodsToDelete = append(oldPodsToDelete, pod.Name)
	}

	glog.V(4).Infof("Marking old pods for deletion")
	for _, pod := range oldAvailablePods {
		if numUnavailable >= maxUnavailable {
			glog.V(4).Infof("Number of unavailable DaemonSet pods: %d, is equal to or exceeds allowed maximum: %d", numUnavailable, maxUnavailable)
			break
		}
		glog.V(4).Infof("Marking pod %s/%s for deletion", ds.Name, pod.Name)
		oldPodsToDelete = append(oldPodsToDelete, pod.Name)
		numUnavailable++
	}
	return dsc.syncNodes(ds, oldPodsToDelete, []string{}, hash)
}

func (dsc *DaemonSetsController) constructHistory(ds *extensions.DaemonSet) (cur *apps.ControllerRevision, old []*apps.ControllerRevision, err error) {
	var histories []*apps.ControllerRevision
	var currentHistories []*apps.ControllerRevision
	histories, err = dsc.controlledHistories(ds)
	if err != nil {
		return nil, nil, err
	}
	for _, history := range histories {
		// Add the unique label if it's not already added to the history
		// We use history name instead of computing hash, so that we don't need to worry about hash collision
		if _, ok := history.Labels[extensions.DefaultDaemonSetUniqueLabelKey]; !ok {
			var clone interface{}
			clone, err = api.Scheme.DeepCopy(history)
			if err != nil {
				return nil, nil, err
			}
			toUpdate := clone.(*apps.ControllerRevision)
			toUpdate.Labels[extensions.DefaultDaemonSetUniqueLabelKey] = toUpdate.Name
			history, err = dsc.kubeClient.AppsV1beta1().ControllerRevisions(ds.Namespace).Update(toUpdate)
			if err != nil {
				return nil, nil, err
			}
		}
		// Compare histories with ds to separate cur and old history
		found := false
		found, err = Match(&ds.Spec.Template, history)
		if err != nil {
			return nil, nil, err
		}
		if found {
			currentHistories = append(currentHistories, history)
		} else {
			old = append(old, history)
		}
	}

	currRevision := maxRevision(old) + 1
	switch len(currentHistories) {
	case 0:
		// Create a new history if the current one isn't found
		cur, err = dsc.snapshot(ds, currRevision)
		if err != nil {
			return nil, nil, err
		}
	default:
		cur, err = dsc.dedupCurHistories(ds, currentHistories)
		if err != nil {
			return nil, nil, err
		}
		// Update revision number if necessary
		if cur.Revision < currRevision {
			var clone interface{}
			clone, err = api.Scheme.DeepCopy(cur)
			if err != nil {
				return nil, nil, err
			}
			toUpdate := clone.(*apps.ControllerRevision)
			toUpdate.Revision = currRevision
			_, err = dsc.kubeClient.AppsV1beta1().ControllerRevisions(ds.Namespace).Update(toUpdate)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	// Label ds with current history's unique label as well
	if ds.Labels[extensions.DefaultDaemonSetUniqueLabelKey] != cur.Labels[extensions.DefaultDaemonSetUniqueLabelKey] {
		var clone interface{}
		clone, err = api.Scheme.DeepCopy(ds)
		if err != nil {
			return nil, nil, err
		}
		toUpdate := clone.(*extensions.DaemonSet)
		if toUpdate.Labels == nil {
			toUpdate.Labels = make(map[string]string)
		}
		toUpdate.Labels[extensions.DefaultDaemonSetUniqueLabelKey] = cur.Labels[extensions.DefaultDaemonSetUniqueLabelKey]
		_, err = dsc.kubeClient.ExtensionsV1beta1().DaemonSets(ds.Namespace).UpdateStatus(toUpdate)
	}
	return cur, old, err
}

func (dsc *DaemonSetsController) cleanupHistory(ds *extensions.DaemonSet) error {
	nodesToDaemonPods, err := dsc.getNodesToDaemonPods(ds)
	if err != nil {
		return fmt.Errorf("couldn't get node to daemon pod mapping for daemon set %q: %v", ds.Name, err)
	}
	_, old, err := dsc.constructHistory(ds)
	if err != nil {
		return fmt.Errorf("failed to construct revisions of DaemonSet: %v", err)
	}

	toKeep := int(*ds.Spec.RevisionHistoryLimit)
	toKill := len(old) - toKeep
	if toKill <= 0 {
		return nil
	}

	// Find all hashes of live pods
	liveHashes := make(map[string]bool)
	for _, pods := range nodesToDaemonPods {
		for _, pod := range pods {
			if hash := pod.Labels[extensions.DefaultDaemonSetUniqueLabelKey]; len(hash) > 0 {
				liveHashes[hash] = true
			}
		}
	}

	// Find all live history with the above hashes
	liveHistory := make(map[string]bool)
	for _, history := range old {
		if hash := history.Labels[extensions.DefaultDaemonSetUniqueLabelKey]; liveHashes[hash] {
			liveHistory[history.Name] = true
		}
	}

	// Clean up old history from smallest to highest revision (from oldest to newest)
	sort.Sort(historiesByRevision(old))
	for _, history := range old {
		if toKill <= 0 {
			break
		}
		if liveHistory[history.Name] {
			continue
		}
		// Clean up
		err := dsc.kubeClient.AppsV1beta1().ControllerRevisions(ds.Namespace).Delete(history.Name, nil)
		if err != nil {
			return err
		}
		toKill--
	}
	return nil
}

// maxRevision returns the max revision number of the given list of histories
func maxRevision(histories []*apps.ControllerRevision) int64 {
	max := int64(0)
	for _, history := range histories {
		if history.Revision > max {
			max = history.Revision
		}
	}
	return max
}

func (dsc *DaemonSetsController) dedupCurHistories(ds *extensions.DaemonSet, curHistories []*apps.ControllerRevision) (*apps.ControllerRevision, error) {
	if len(curHistories) == 1 {
		return curHistories[0], nil
	}
	var maxRevision int64
	var keepCur *apps.ControllerRevision
	for _, cur := range curHistories {
		if cur.Revision >= maxRevision {
			keepCur = cur
			maxRevision = cur.Revision
		}
	}
	// Clean up duplicates and relabel pods
	for _, cur := range curHistories {
		if cur.Name == keepCur.Name {
			continue
		}
		// Relabel pods before dedup
		pods, err := dsc.getDaemonPods(ds)
		if err != nil {
			return nil, err
		}
		for _, pod := range pods {
			if pod.Labels[extensions.DefaultDaemonSetUniqueLabelKey] != keepCur.Labels[extensions.DefaultDaemonSetUniqueLabelKey] {
				clone, err := api.Scheme.DeepCopy(pod)
				if err != nil {
					return nil, err
				}
				toUpdate := clone.(*v1.Pod)
				if toUpdate.Labels == nil {
					toUpdate.Labels = make(map[string]string)
				}
				toUpdate.Labels[extensions.DefaultDaemonSetUniqueLabelKey] = keepCur.Labels[extensions.DefaultDaemonSetUniqueLabelKey]
				_, err = dsc.kubeClient.Core().Pods(ds.Namespace).Update(toUpdate)
				if err != nil {
					return nil, err
				}
			}
		}
		// Remove duplicates
		err = dsc.kubeClient.AppsV1beta1().ControllerRevisions(ds.Namespace).Delete(cur.Name, nil)
		if err != nil {
			return nil, err
		}
	}
	return keepCur, nil
}

// controlledHistories returns all ControllerRevisions controlled by the given DaemonSet
// Note that returned histories are pointers to objects in the cache.
// If you want to modify one, you need to deep-copy it first.
func (dsc *DaemonSetsController) controlledHistories(ds *extensions.DaemonSet) ([]*apps.ControllerRevision, error) {
	var result []*apps.ControllerRevision
	selector, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		return nil, err
	}
	histories, err := dsc.historyLister.List(selector)
	if err != nil {
		return nil, err
	}
	for _, history := range histories {
		// Skip history that doesn't belong to the DaemonSet
		if controllerRef := controller.GetControllerOf(history); controllerRef == nil || controllerRef.UID != ds.UID {
			continue
		}
		result = append(result, history)
	}
	return result, nil
}

// Match check if ds template is semantically equal to the template stored in history
func Match(template *v1.PodTemplateSpec, history *apps.ControllerRevision) (bool, error) {
	t, err := DecodeHistory(history)
	return apiequality.Semantic.DeepEqual(template, t), err
}

func DecodeHistory(history *apps.ControllerRevision) (*v1.PodTemplateSpec, error) {
	template := v1.PodTemplateSpec{}
	err := json.Unmarshal(history.Data.Raw, &template)
	return &template, err
}

func encodeTemplate(template *v1.PodTemplateSpec) ([]byte, error) {
	return json.Marshal(template)
}

func (dsc *DaemonSetsController) snapshot(ds *extensions.DaemonSet, revision int64) (*apps.ControllerRevision, error) {
	encodedTemplate, err := encodeTemplate(&ds.Spec.Template)
	if err != nil {
		return nil, err
	}
	hash := fmt.Sprint(controller.ComputeHash(&ds.Spec.Template, ds.Status.CollisionCount))
	name := ds.Name + "-" + hash
	history := &apps.ControllerRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       ds.Namespace,
			Labels:          labelsutil.CloneAndAddLabel(ds.Spec.Template.Labels, extensions.DefaultDaemonSetUniqueLabelKey, hash),
			Annotations:     ds.Annotations,
			OwnerReferences: []metav1.OwnerReference{*newControllerRef(ds)},
		},
		Data:     runtime.RawExtension{Raw: encodedTemplate},
		Revision: revision,
	}

	history, err = dsc.kubeClient.AppsV1beta1().ControllerRevisions(ds.Namespace).Create(history)
	if errors.IsAlreadyExists(err) {
		// TODO: Is it okay to get from historyLister?
		existedHistory, getErr := dsc.kubeClient.AppsV1beta1().ControllerRevisions(ds.Namespace).Get(name, metav1.GetOptions{})
		if getErr != nil {
			return nil, getErr
		}
		// Check if we already created it
		done, err := Match(&ds.Spec.Template, existedHistory)
		if err != nil {
			return nil, err
		}
		if done {
			return existedHistory, nil
		}

		// Handle name collisions between different history
		// TODO: Is it okay to get from dsLister?
		currDS, getErr := dsc.kubeClient.ExtensionsV1beta1().DaemonSets(ds.Namespace).Get(ds.Name, metav1.GetOptions{})
		if getErr != nil {
			return nil, getErr
		}
		if currDS.Status.CollisionCount == nil {
			currDS.Status.CollisionCount = new(int64)
		}
		*currDS.Status.CollisionCount++
		_, updateErr := dsc.kubeClient.ExtensionsV1beta1().DaemonSets(ds.Namespace).UpdateStatus(currDS)
		if updateErr != nil {
			return nil, updateErr
		}
		glog.V(2).Infof("Found a hash collision for DaemonSet %q - bumping collisionCount to %d to resolve it", ds.Name, *currDS.Status.CollisionCount)
		return nil, err
	}
	return history, err
}

func (dsc *DaemonSetsController) getAllDaemonSetPods(ds *extensions.DaemonSet, nodeToDaemonPods map[string][]*v1.Pod, hash string) ([]*v1.Pod, []*v1.Pod) {
	var newPods []*v1.Pod
	var oldPods []*v1.Pod

	for _, pods := range nodeToDaemonPods {
		for _, pod := range pods {
			if util.IsPodUpdated(ds.Spec.TemplateGeneration, pod, hash) {
				newPods = append(newPods, pod)
			} else {
				oldPods = append(oldPods, pod)
			}
		}
	}
	return newPods, oldPods
}

func (dsc *DaemonSetsController) getUnavailableNumbers(ds *extensions.DaemonSet, nodeToDaemonPods map[string][]*v1.Pod) (int, int, error) {
	glog.V(4).Infof("Getting unavailable numbers")
	// TODO: get nodeList once in syncDaemonSet and pass it to other functions
	nodeList, err := dsc.nodeLister.List(labels.Everything())
	if err != nil {
		return -1, -1, fmt.Errorf("couldn't get list of nodes during rolling update of daemon set %#v: %v", ds, err)
	}

	var numUnavailable, desiredNumberScheduled int
	for i := range nodeList {
		node := nodeList[i]
		wantToRun, _, _, err := dsc.nodeShouldRunDaemonPod(node, ds)
		if err != nil {
			return -1, -1, err
		}
		if !wantToRun {
			continue
		}
		desiredNumberScheduled++
		daemonPods, exists := nodeToDaemonPods[node.Name]
		if !exists {
			numUnavailable++
			continue
		}
		available := false
		for _, pod := range daemonPods {
			if podutil.IsPodAvailable(pod, ds.Spec.MinReadySeconds, metav1.Now()) {
				available = true
				break
			}
		}
		if !available {
			numUnavailable++
		}
	}
	maxUnavailable, err := intstrutil.GetValueFromIntOrPercent(ds.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable, desiredNumberScheduled, true)
	if err != nil {
		return -1, -1, fmt.Errorf("Invalid value for MaxUnavailable: %v", err)
	}
	glog.V(4).Infof(" DaemonSet %s/%s, maxUnavailable: %d, numUnavailable: %d", ds.Namespace, ds.Name, maxUnavailable, numUnavailable)
	return maxUnavailable, numUnavailable, nil
}

type historiesByRevision []*apps.ControllerRevision

func (h historiesByRevision) Len() int      { return len(h) }
func (h historiesByRevision) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h historiesByRevision) Less(i, j int) bool {
	return h[i].Revision < h[j].Revision
}
