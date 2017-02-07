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

package app

import (
	"fmt"
	"io/ioutil"
	"os"

	"k8s.io/kubernetes/plugin/cmd/kube-scheduler/app/options"

	"k8s.io/apimachinery/pkg/runtime"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"

	clientv1 "k8s.io/client-go/pkg/api/v1"

	"k8s.io/kubernetes/plugin/pkg/scheduler"
	_ "k8s.io/kubernetes/plugin/pkg/scheduler/algorithmprovider"
	schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
	latestschedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api/latest"
	"k8s.io/kubernetes/plugin/pkg/scheduler/factory"

	"github.com/golang/glog"
)

func createRecorder(kubecli *clientset.Clientset, s *options.SchedulerServer) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubecli.Core().RESTClient()).Events("")})
	return eventBroadcaster.NewRecorder(api.Scheme, clientv1.EventSource{Component: s.SchedulerName})
}

func createClient(s *options.SchedulerServer) (*clientset.Clientset, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags(s.Master, s.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("unable to build config from flags: %v", err)
	}

	kubeconfig.ContentType = s.ContentType
	// Override kubeconfig qps/burst settings from flags
	kubeconfig.QPS = s.KubeAPIQPS
	kubeconfig.Burst = int(s.KubeAPIBurst)

	cli, err := clientset.NewForConfig(restclient.AddUserAgent(kubeconfig, "leader-election"))
	if err != nil {
		return nil, fmt.Errorf("invalid API configuration: %v", err)
	}
	return cli, nil
}

// createScheduler encapsulates the entire creation of a runnable scheduler.
func createScheduler(s *options.SchedulerServer, kubecli *clientset.Clientset, recorder record.EventRecorder) (*scheduler.Scheduler, error) {
	configurator := factory.NewConfigFactory(kubecli, s.SchedulerName, s.HardPodAffinitySymmetricWeight, s.FailureDomains)

	// Rebuild the configurator with a default Create(...) method.
	configurator = &schedulerConfigurator{
		configurator,
		s.PolicyConfigFile,
		s.AlgorithmProvider}

	return scheduler.NewFromConfigurator(configurator, func(cfg *scheduler.Config) {
		cfg.Recorder = recorder
	})
}

// schedulerConfigurator is an interface wrapper that provides default Configuration creation based on user
// provided config file.
type schedulerConfigurator struct {
	scheduler.Configurator
	policyFile        string
	algorithmProvider string
}

// Create implements the interface for the Configurator, hence it is exported even through the struct is not.
func (sc schedulerConfigurator) Create() (*scheduler.Config, error) {
	if _, err := os.Stat(sc.policyFile); err != nil {
		if sc.Configurator != nil {
			return sc.Configurator.CreateFromProvider(sc.algorithmProvider)
		}
		return nil, fmt.Errorf("Configurator was nil")
	}

	// policy file is valid, try to create a configuration from it.
	var policy schedulerapi.Policy
	configData, err := ioutil.ReadFile(sc.policyFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read policy config: %v", err)
	}
	if err := runtime.DecodeInto(latestschedulerapi.Codec, configData, &policy); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}
	return sc.CreateFromConfig(policy)
}
