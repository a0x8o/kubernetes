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

package certificates

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	certificates "k8s.io/kubernetes/pkg/apis/certificates/v1beta1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	"k8s.io/kubernetes/pkg/client/legacylisters"
	"k8s.io/kubernetes/pkg/controller"

	"github.com/golang/glog"
)

type AutoApprover interface {
	AutoApprove(csr *certificates.CertificateSigningRequest) (*certificates.CertificateSigningRequest, error)
}

type Signer interface {
	Sign(csr *certificates.CertificateSigningRequest) ([]byte, error)
}

type CertificateController struct {
	kubeClient clientset.Interface

	// CSR framework and store
	csrController cache.Controller
	csrStore      listers.StoreToCertificateRequestLister

	syncHandler func(csrKey string) error

	approver AutoApprover
	signer   Signer

	queue workqueue.RateLimitingInterface
}

func NewCertificateController(kubeClient clientset.Interface, syncPeriod time.Duration, caCertFile, caKeyFile string, approver AutoApprover) (*CertificateController, error) {
	// Send events to the apiserver
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubeClient.Core().RESTClient()).Events("")})

	s, err := NewCFSSLSigner(caCertFile, caKeyFile)
	if err != nil {
		return nil, err
	}

	cc := &CertificateController{
		kubeClient: kubeClient,
		queue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "certificate"),
		signer:     s,
		approver:   approver,
	}

	// Manage the addition/update of certificate requests
	cc.csrStore.Store, cc.csrController = cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return cc.kubeClient.Certificates().CertificateSigningRequests().List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return cc.kubeClient.Certificates().CertificateSigningRequests().Watch(options)
			},
		},
		&certificates.CertificateSigningRequest{},
		syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				csr := obj.(*certificates.CertificateSigningRequest)
				glog.V(4).Infof("Adding certificate request %s", csr.Name)
				cc.enqueueCertificateRequest(obj)
			},
			UpdateFunc: func(old, new interface{}) {
				oldCSR := old.(*certificates.CertificateSigningRequest)
				glog.V(4).Infof("Updating certificate request %s", oldCSR.Name)
				cc.enqueueCertificateRequest(new)
			},
			DeleteFunc: func(obj interface{}) {
				csr, ok := obj.(*certificates.CertificateSigningRequest)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						glog.V(2).Infof("Couldn't get object from tombstone %#v", obj)
						return
					}
					csr, ok = tombstone.Obj.(*certificates.CertificateSigningRequest)
					if !ok {
						glog.V(2).Infof("Tombstone contained object that is not a CSR: %#v", obj)
						return
					}
				}
				glog.V(4).Infof("Deleting certificate request %s", csr.Name)
				cc.enqueueCertificateRequest(obj)
			},
		},
	)
	cc.syncHandler = cc.maybeSignCertificate
	return cc, nil
}

// Run the main goroutine responsible for watching and syncing jobs.
func (cc *CertificateController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer cc.queue.ShutDown()

	go cc.csrController.Run(stopCh)

	glog.Infof("Starting certificate controller manager")
	for i := 0; i < workers; i++ {
		go wait.Until(cc.worker, time.Second, stopCh)
	}
	<-stopCh
	glog.Infof("Shutting down certificate controller")
}

// worker runs a thread that dequeues CSRs, handles them, and marks them done.
func (cc *CertificateController) worker() {
	for cc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false when it's time to quit.
func (cc *CertificateController) processNextWorkItem() bool {
	cKey, quit := cc.queue.Get()
	if quit {
		return false
	}
	defer cc.queue.Done(cKey)

	err := cc.syncHandler(cKey.(string))
	if err == nil {
		cc.queue.Forget(cKey)
		return true
	}

	cc.queue.AddRateLimited(cKey)
	utilruntime.HandleError(fmt.Errorf("Sync %v failed with : %v", cKey, err))
	return true
}

func (cc *CertificateController) enqueueCertificateRequest(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %+v: %v", obj, err))
		return
	}
	cc.queue.Add(key)
}

// maybeSignCertificate will inspect the certificate request and, if it has
// been approved and meets policy expectations, generate an X509 cert using the
// cluster CA assets. If successful it will update the CSR approve subresource
// with the signed certificate.
func (cc *CertificateController) maybeSignCertificate(key string) error {
	startTime := time.Now()
	defer func() {
		glog.V(4).Infof("Finished syncing certificate request %q (%v)", key, time.Now().Sub(startTime))
	}()
	obj, exists, err := cc.csrStore.Store.GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		glog.V(3).Infof("csr has been deleted: %v", key)
		return nil
	}
	csr := obj.(*certificates.CertificateSigningRequest)

	if cc.approver != nil {
		csr, err = cc.approver.AutoApprove(csr)
		if err != nil {
			return fmt.Errorf("error auto approving csr: %v", err)
		}
	}

	// At this point, the controller needs to:
	// 1. Check the approval conditions
	// 2. Generate a signed certificate
	// 3. Update the Status subresource

	if csr.Status.Certificate == nil && IsCertificateRequestApproved(csr) {
		certBytes, err := cc.signer.Sign(csr)
		if err != nil {
			return err
		}
		csr.Status.Certificate = certBytes
	}

	_, err = cc.kubeClient.Certificates().CertificateSigningRequests().UpdateStatus(csr)
	return err
}
