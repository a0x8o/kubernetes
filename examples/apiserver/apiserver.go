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

package apiserver

import (
	"fmt"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/kubernetes/cmd/libs/go2idl/client-gen/test_apis/testgroup/v1"
	testgroupetcd "k8s.io/kubernetes/examples/apiserver/rest"
	"k8s.io/kubernetes/pkg/api"
	kubeoptions "k8s.io/kubernetes/pkg/kubeapiserver/options"

	// Install the testgroup API
	_ "k8s.io/kubernetes/cmd/libs/go2idl/client-gen/test_apis/testgroup/install"
)

const (
	// Ports on which to run the server.
	// Explicitly setting these to a different value than the default values, to prevent this from clashing with a local cluster.
	InsecurePort = 8081
	SecurePort   = 6444
)

type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions
	Etcd                    *genericoptions.EtcdOptions
	SecureServing           *genericoptions.SecureServingOptions
	InsecureServing         *genericoptions.ServingOptions
	Authentication          *kubeoptions.BuiltInAuthenticationOptions
	CloudProvider           *kubeoptions.CloudProviderOptions
}

func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		Etcd:            genericoptions.NewEtcdOptions(storagebackend.NewDefaultConfig(kubeoptions.DefaultEtcdPathPrefix, api.Scheme, nil)),
		SecureServing:   genericoptions.NewSecureServingOptions(),
		InsecureServing: genericoptions.NewInsecureServingOptions(),
		Authentication:  kubeoptions.NewBuiltInAuthenticationOptions().WithAll(),
		CloudProvider:   kubeoptions.NewCloudProviderOptions(),
	}
	s.InsecureServing.BindPort = InsecurePort
	s.SecureServing.ServingOptions.BindPort = SecurePort
	s.Etcd.StorageConfig.ServerList = []string{"http://127.0.0.1:2379"}

	return &s
}

func (serverOptions *ServerRunOptions) Run(stopCh <-chan struct{}) error {
	serverOptions.Etcd.StorageConfig.ServerList = []string{"http://127.0.0.1:2379"}

	// set defaults
	if err := serverOptions.CloudProvider.DefaultExternalHost(serverOptions.GenericServerRunOptions); err != nil {
		return err
	}
	if err := serverOptions.SecureServing.MaybeDefaultWithSelfSignedCerts(serverOptions.GenericServerRunOptions.AdvertiseAddress.String()); err != nil {
		glog.Fatalf("Error creating self-signed certificates: %v", err)
	}

	// validate options
	if errs := serverOptions.Etcd.Validate(); len(errs) > 0 {
		return utilerrors.NewAggregate(errs)
	}
	if errs := serverOptions.SecureServing.Validate(); len(errs) > 0 {
		return utilerrors.NewAggregate(errs)
	}
	if errs := serverOptions.InsecureServing.Validate("insecure-port"); len(errs) > 0 {
		return utilerrors.NewAggregate(errs)
	}

	// create config from options
	config := genericapiserver.NewConfig().
		WithSerializer(api.Codecs)

	if err := serverOptions.GenericServerRunOptions.ApplyTo(config); err != nil {
		return err
	}
	if err := serverOptions.InsecureServing.ApplyTo(config); err != nil {
		return err
	}
	if err := serverOptions.SecureServing.ApplyTo(config); err != nil {
		return fmt.Errorf("failed to configure https: %s", err)
	}
	if err := serverOptions.Authentication.ApplyTo(config); err != nil {
		return fmt.Errorf("failed to configure authentication: %s", err)
	}

	config.Authorizer = authorizerfactory.NewAlwaysAllowAuthorizer()
	config.SwaggerConfig = genericapiserver.DefaultSwaggerConfig()

	groupVersion := v1.SchemeGroupVersion
	groupName := groupVersion.Group
	groupMeta, err := api.Registry.Group(groupName)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	storageFactory := serverstorage.NewDefaultStorageFactory(serverOptions.Etcd.StorageConfig, "application/json", api.Codecs, serverstorage.NewDefaultResourceEncodingConfig(api.Registry), serverstorage.NewResourceConfig())
	storageConfig, err := storageFactory.NewConfig(schema.GroupResource{Group: groupName, Resource: "testtype"})
	if err != nil {
		return fmt.Errorf("Unable to get storage config: %v", err)
	}
	if err := serverOptions.Etcd.ApplyWithStorageFactoryTo(storageFactory, config); err != nil {
		return fmt.Errorf("failed to configure authentication: %s", err)
	}

	s, err := config.Complete().New()
	if err != nil {
		return fmt.Errorf("Error in bringing up the server: %v", err)
	}

	testTypeOpts := generic.RESTOptions{
		StorageConfig:           storageConfig,
		Decorator:               generic.UndecoratedStorage,
		ResourcePrefix:          "testtypes",
		DeleteCollectionWorkers: 1,
	}

	restStorageMap := map[string]rest.Storage{
		"testtypes": testgroupetcd.NewREST(testTypeOpts),
	}
	apiGroupInfo := genericapiserver.APIGroupInfo{
		GroupMeta: *groupMeta,
		VersionedResourcesStorageMap: map[string]map[string]rest.Storage{
			groupVersion.Version: restStorageMap,
		},
		Scheme:               api.Scheme,
		NegotiatedSerializer: api.Codecs,
	}
	if err := s.InstallAPIGroup(&apiGroupInfo); err != nil {
		return fmt.Errorf("Error in installing API: %v", err)
	}
	s.PrepareRun().Run(stopCh)
	return nil
}
