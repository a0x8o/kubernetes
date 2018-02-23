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

package testing

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/categories"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi"
	openapitesting "k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi/testing"
	"k8s.io/kubernetes/pkg/kubectl/plugins"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
	"k8s.io/kubernetes/pkg/kubectl/validation"
	"k8s.io/kubernetes/pkg/printers"
	metricsclientset "k8s.io/metrics/pkg/client/clientset_generated/clientset"
)

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type InternalType struct {
	Kind       string
	APIVersion string

	Name string
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ExternalType struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`

	Name string `json:"name"`
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ExternalType2 struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`

	Name string `json:"name"`
}

func (obj *InternalType) GetObjectKind() schema.ObjectKind { return obj }
func (obj *InternalType) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}
func (obj *InternalType) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}
func (obj *ExternalType) GetObjectKind() schema.ObjectKind { return obj }
func (obj *ExternalType) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}
func (obj *ExternalType) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}
func (obj *ExternalType2) GetObjectKind() schema.ObjectKind { return obj }
func (obj *ExternalType2) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}
func (obj *ExternalType2) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}

func NewInternalType(kind, apiversion, name string) *InternalType {
	item := InternalType{Kind: kind,
		APIVersion: apiversion,
		Name:       name}
	return &item
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type InternalNamespacedType struct {
	Kind       string
	APIVersion string

	Name      string
	Namespace string
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ExternalNamespacedType struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`

	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ExternalNamespacedType2 struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`

	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func (obj *InternalNamespacedType) GetObjectKind() schema.ObjectKind { return obj }
func (obj *InternalNamespacedType) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}
func (obj *InternalNamespacedType) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}
func (obj *ExternalNamespacedType) GetObjectKind() schema.ObjectKind { return obj }
func (obj *ExternalNamespacedType) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}
func (obj *ExternalNamespacedType) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}
func (obj *ExternalNamespacedType2) GetObjectKind() schema.ObjectKind { return obj }
func (obj *ExternalNamespacedType2) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}
func (obj *ExternalNamespacedType2) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}

func NewInternalNamespacedType(kind, apiversion, name, namespace string) *InternalNamespacedType {
	item := InternalNamespacedType{Kind: kind,
		APIVersion: apiversion,
		Name:       name,
		Namespace:  namespace}
	return &item
}

var versionErr = errors.New("not a version")

func versionErrIfFalse(b bool) error {
	if b {
		return nil
	}
	return versionErr
}

var ValidVersion = legacyscheme.Registry.GroupOrDie(api.GroupName).GroupVersion.Version
var InternalGV = schema.GroupVersion{Group: "apitest", Version: runtime.APIVersionInternal}
var UnlikelyGV = schema.GroupVersion{Group: "apitest", Version: "unlikelyversion"}
var ValidVersionGV = schema.GroupVersion{Group: "apitest", Version: ValidVersion}

func NewExternalScheme() (*runtime.Scheme, meta.RESTMapper, runtime.Codec) {
	scheme := runtime.NewScheme()
	mapper, codec := AddToScheme(scheme)
	return scheme, mapper, codec
}

func AddToScheme(scheme *runtime.Scheme) (meta.RESTMapper, runtime.Codec) {
	scheme.AddKnownTypeWithName(InternalGV.WithKind("Type"), &InternalType{})
	scheme.AddKnownTypeWithName(UnlikelyGV.WithKind("Type"), &ExternalType{})
	//This tests that kubectl will not confuse the external scheme with the internal scheme, even when they accidentally have versions of the same name.
	scheme.AddKnownTypeWithName(ValidVersionGV.WithKind("Type"), &ExternalType2{})

	scheme.AddKnownTypeWithName(InternalGV.WithKind("NamespacedType"), &InternalNamespacedType{})
	scheme.AddKnownTypeWithName(UnlikelyGV.WithKind("NamespacedType"), &ExternalNamespacedType{})
	//This tests that kubectl will not confuse the external scheme with the internal scheme, even when they accidentally have versions of the same name.
	scheme.AddKnownTypeWithName(ValidVersionGV.WithKind("NamespacedType"), &ExternalNamespacedType2{})

	codecs := serializer.NewCodecFactory(scheme)
	codec := codecs.LegacyCodec(UnlikelyGV)
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{UnlikelyGV, ValidVersionGV}, func(version schema.GroupVersion) (*meta.VersionInterfaces, error) {
		return &meta.VersionInterfaces{
			ObjectConvertor:  scheme,
			MetadataAccessor: meta.NewAccessor(),
		}, versionErrIfFalse(version == ValidVersionGV || version == UnlikelyGV)
	})
	for _, gv := range []schema.GroupVersion{UnlikelyGV, ValidVersionGV} {
		for kind := range scheme.KnownTypes(gv) {
			gvk := gv.WithKind(kind)

			scope := meta.RESTScopeNamespace
			mapper.Add(gvk, scope)
		}
	}

	return mapper, codec
}

type fakeCachedDiscoveryClient struct {
	discovery.DiscoveryInterface
}

func (d *fakeCachedDiscoveryClient) Fresh() bool {
	return true
}

func (d *fakeCachedDiscoveryClient) Invalidate() {
}

func (d *fakeCachedDiscoveryClient) ServerResources() ([]*metav1.APIResourceList, error) {
	return []*metav1.APIResourceList{}, nil
}

type TestFactory struct {
	Client             kubectl.RESTClient
	UnstructuredClient kubectl.RESTClient
	Describer          printers.Describer
	Validator          validation.Schema
	Namespace          string
	ClientConfig       *restclient.Config
	Err                error
	Command            string
	TmpDir             string
	CategoryExpander   categories.CategoryExpander
	MetricsClientSet   metricsclientset.Interface

	ClientForMappingFunc             func(mapping *meta.RESTMapping) (resource.RESTClient, error)
	UnstructuredClientForMappingFunc func(mapping *meta.RESTMapping) (resource.RESTClient, error)
	OpenAPISchemaFunc                func() (openapi.Resources, error)
}

type FakeFactory struct {
	tf *TestFactory
}

func NewTestFactory() (cmdutil.Factory, *TestFactory) {
	t := &TestFactory{
		Validator: validation.NullSchema{},
	}
	return &FakeFactory{
		tf: t,
	}, t
}

func (f *FakeFactory) DiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(f.tf.ClientConfig)
	if err != nil {
		return nil, err
	}
	return &fakeCachedDiscoveryClient{DiscoveryInterface: discoveryClient}, nil
}

func (f *FakeFactory) FlagSet() *pflag.FlagSet {
	return nil
}

func (f *FakeFactory) Object() (meta.RESTMapper, runtime.ObjectTyper) {
	groupResources := testDynamicResources()
	mapper := discovery.NewRESTMapper(groupResources, meta.InterfacesForUnstructuredConversion(legacyscheme.Registry.InterfacesFor))
	typer := discovery.NewUnstructuredObjectTyper(groupResources, legacyscheme.Scheme)

	fakeDs := &fakeCachedDiscoveryClient{}
	expander := cmdutil.NewShortcutExpander(mapper, fakeDs)
	return expander, typer
}

func (f *FakeFactory) CategoryExpander() categories.CategoryExpander {
	return categories.LegacyCategoryExpander
}

func (f *FakeFactory) RESTClient() (*restclient.RESTClient, error) {
	return nil, nil
}

func (f *FakeFactory) KubernetesClientSet() (*kubernetes.Clientset, error) {
	return nil, nil
}

func (f *FakeFactory) MetricsClientSet() (metricsclientset.Interface, error) {
	return f.tf.MetricsClientSet, f.tf.Err
}

func (f *FakeFactory) ClientSet() (internalclientset.Interface, error) {
	return nil, nil
}

func (f *FakeFactory) ClientConfig() (*restclient.Config, error) {
	return f.tf.ClientConfig, f.tf.Err
}

func (f *FakeFactory) BareClientConfig() (*restclient.Config, error) {
	return f.tf.ClientConfig, f.tf.Err
}

func (f *FakeFactory) ClientForMapping(mapping *meta.RESTMapping) (resource.RESTClient, error) {
	if f.tf.ClientForMappingFunc != nil {
		return f.tf.ClientForMappingFunc(mapping)
	}
	return f.tf.Client, f.tf.Err
}

func (f *FakeFactory) ClientSetForVersion(requiredVersion *schema.GroupVersion) (internalclientset.Interface, error) {
	return nil, nil
}
func (f *FakeFactory) ClientConfigForVersion(requiredVersion *schema.GroupVersion) (*restclient.Config, error) {
	return nil, nil
}

func (f *FakeFactory) UnstructuredClientForMapping(mapping *meta.RESTMapping) (resource.RESTClient, error) {
	if f.tf.UnstructuredClientForMappingFunc != nil {
		return f.tf.UnstructuredClientForMappingFunc(mapping)
	}
	return f.tf.UnstructuredClient, f.tf.Err
}

func (f *FakeFactory) Describer(*meta.RESTMapping) (printers.Describer, error) {
	return f.tf.Describer, f.tf.Err
}

func (f *FakeFactory) Scaler(*meta.RESTMapping) (kubectl.Scaler, error) {
	return nil, nil
}

func (f *FakeFactory) Reaper(*meta.RESTMapping) (kubectl.Reaper, error) {
	return nil, nil
}

func (f *FakeFactory) HistoryViewer(*meta.RESTMapping) (kubectl.HistoryViewer, error) {
	return nil, nil
}

func (f *FakeFactory) Rollbacker(*meta.RESTMapping) (kubectl.Rollbacker, error) {
	return nil, nil
}

func (f *FakeFactory) StatusViewer(*meta.RESTMapping) (kubectl.StatusViewer, error) {
	return nil, nil
}

func (f *FakeFactory) MapBasedSelectorForObject(runtime.Object) (string, error) {
	return "", nil
}

func (f *FakeFactory) PortsForObject(runtime.Object) ([]string, error) {
	return nil, nil
}

func (f *FakeFactory) ProtocolsForObject(runtime.Object) (map[string]string, error) {
	return nil, nil
}

func (f *FakeFactory) LabelsForObject(runtime.Object) (map[string]string, error) {
	return nil, nil
}

func (f *FakeFactory) LogsForObject(object, options runtime.Object, timeout time.Duration) (*restclient.Request, error) {
	return nil, nil
}

func (f *FakeFactory) Pauser(info *resource.Info) ([]byte, error) {
	return nil, nil
}

func (f *FakeFactory) Resumer(info *resource.Info) ([]byte, error) {
	return nil, nil
}

func (f *FakeFactory) ResolveImage(name string) (string, error) {
	return name, nil
}

func (f *FakeFactory) Validator(validate bool) (validation.Schema, error) {
	return f.tf.Validator, f.tf.Err
}

func (f *FakeFactory) OpenAPISchema() (openapi.Resources, error) {
	return nil, nil
}

func (f *FakeFactory) DefaultNamespace() (string, bool, error) {
	return f.tf.Namespace, false, f.tf.Err
}

func (f *FakeFactory) Generators(cmdName string) map[string]kubectl.Generator {
	var generator map[string]kubectl.Generator
	switch cmdName {
	case "run":
		generator = map[string]kubectl.Generator{
			cmdutil.DeploymentV1Beta1GeneratorName: kubectl.DeploymentV1Beta1{},
		}
	}
	return generator
}

func (f *FakeFactory) CanBeExposed(schema.GroupKind) error {
	return nil
}

func (f *FakeFactory) CanBeAutoscaled(schema.GroupKind) error {
	return nil
}

func (f *FakeFactory) AttachablePodForObject(ob runtime.Object, timeout time.Duration) (*api.Pod, error) {
	return nil, nil
}

func (f *FakeFactory) ApproximatePodTemplateForObject(obj runtime.Object) (*api.PodTemplateSpec, error) {
	return f.ApproximatePodTemplateForObject(obj)
}

func (f *FakeFactory) UpdatePodSpecForObject(obj runtime.Object, fn func(*v1.PodSpec) error) (bool, error) {
	return false, nil
}

func (f *FakeFactory) EditorEnvs() []string {
	return nil
}

func (f *FakeFactory) Command(*cobra.Command, bool) string {
	return f.tf.Command
}

func (f *FakeFactory) BindFlags(flags *pflag.FlagSet) {
}

func (f *FakeFactory) BindExternalFlags(flags *pflag.FlagSet) {
}

func (f *FakeFactory) NewBuilder() *resource.Builder {
	mapper, typer := f.Object()

	return resource.NewBuilder(
		&resource.Mapper{
			RESTMapper:   mapper,
			ObjectTyper:  typer,
			ClientMapper: resource.ClientMapperFunc(f.ClientForMapping),
			Decoder:      cmdutil.InternalVersionDecoder(),
		},
		&resource.Mapper{
			RESTMapper:   mapper,
			ObjectTyper:  typer,
			ClientMapper: resource.ClientMapperFunc(f.UnstructuredClientForMapping),
			Decoder:      unstructured.UnstructuredJSONScheme,
		},
		f.CategoryExpander(),
	)
}

func (f *FakeFactory) DefaultResourceFilterOptions(cmd *cobra.Command, withNamespace bool) *printers.PrintOptions {
	return &printers.PrintOptions{}
}

func (f *FakeFactory) DefaultResourceFilterFunc() kubectl.Filters {
	return nil
}

func (f *FakeFactory) SuggestedPodTemplateResources() []schema.GroupResource {
	return []schema.GroupResource{}
}

func (f *FakeFactory) PluginLoader() plugins.PluginLoader {
	return &plugins.DummyPluginLoader{}
}

func (f *FakeFactory) PluginRunner() plugins.PluginRunner {
	return &plugins.ExecPluginRunner{}
}

type fakeAPIFactory struct {
	cmdutil.Factory
	tf *TestFactory
}

func (f *fakeAPIFactory) Object() (meta.RESTMapper, runtime.ObjectTyper) {
	groupResources := testDynamicResources()
	mapper := discovery.NewRESTMapper(
		groupResources,
		meta.InterfacesForUnstructuredConversion(func(version schema.GroupVersion) (*meta.VersionInterfaces, error) {
			switch version {
			// provide typed objects for these two versions
			case ValidVersionGV, UnlikelyGV:
				return &meta.VersionInterfaces{
					ObjectConvertor:  scheme.Scheme,
					MetadataAccessor: meta.NewAccessor(),
				}, nil
			// otherwise fall back to the legacy scheme
			default:
				return legacyscheme.Registry.InterfacesFor(version)
			}
		}),
	)
	// for backwards compatibility with existing tests, allow rest mappings from the scheme to show up
	// TODO: make this opt-in?
	mapper = meta.FirstHitRESTMapper{
		MultiRESTMapper: meta.MultiRESTMapper{
			mapper,
			legacyscheme.Registry.RESTMapper(),
		},
	}

	// TODO: should probably be the external scheme
	typer := discovery.NewUnstructuredObjectTyper(groupResources, legacyscheme.Scheme)
	fakeDs := &fakeCachedDiscoveryClient{}
	expander := cmdutil.NewShortcutExpander(mapper, fakeDs)
	return expander, typer
}

func (f *fakeAPIFactory) KubernetesClientSet() (*kubernetes.Clientset, error) {
	fakeClient := f.tf.Client.(*fake.RESTClient)
	clientset := kubernetes.NewForConfigOrDie(f.tf.ClientConfig)

	clientset.CoreV1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.AuthorizationV1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.AuthorizationV1beta1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.AuthorizationV1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.AuthorizationV1beta1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.AutoscalingV1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.AutoscalingV2beta1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.BatchV1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.BatchV2alpha1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.CertificatesV1beta1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.ExtensionsV1beta1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.RbacV1alpha1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.RbacV1beta1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.StorageV1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.StorageV1beta1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.AppsV1beta1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.AppsV1beta2().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.PolicyV1beta1().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.DiscoveryClient.RESTClient().(*restclient.RESTClient).Client = fakeClient.Client

	return clientset, f.tf.Err
}

func (f *fakeAPIFactory) MetricsClientSet() (metricsclientset.Interface, error) {
	return f.tf.MetricsClientSet, f.tf.Err
}

func (f *fakeAPIFactory) ClientSet() (internalclientset.Interface, error) {
	// Swap the HTTP client out of the REST client with the fake
	// version.
	fakeClient := f.tf.Client.(*fake.RESTClient)
	clientset := internalclientset.NewForConfigOrDie(f.tf.ClientConfig)
	clientset.Core().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.Authentication().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.Authorization().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.Autoscaling().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.Batch().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.Certificates().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.Extensions().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.Rbac().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.Storage().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.Apps().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.Policy().RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	clientset.DiscoveryClient.RESTClient().(*restclient.RESTClient).Client = fakeClient.Client
	return clientset, f.tf.Err
}

func (f *fakeAPIFactory) RESTClient() (*restclient.RESTClient, error) {
	// Swap out the HTTP client out of the client with the fake's version.
	fakeClient := f.tf.Client.(*fake.RESTClient)
	restClient, err := restclient.RESTClientFor(f.tf.ClientConfig)
	if err != nil {
		panic(err)
	}
	restClient.Client = fakeClient.Client
	return restClient, f.tf.Err
}

func (f *fakeAPIFactory) DiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	fakeClient := f.tf.Client.(*fake.RESTClient)
	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(f.tf.ClientConfig)
	discoveryClient.RESTClient().(*restclient.RESTClient).Client = fakeClient.Client

	cacheDir := filepath.Join(f.tf.TmpDir, ".kube", "cache", "discovery")
	return cmdutil.NewCachedDiscoveryClient(discoveryClient, cacheDir, time.Duration(10*time.Minute)), nil
}

func (f *fakeAPIFactory) CategoryExpander() categories.CategoryExpander {
	if f.tf.CategoryExpander != nil {
		return f.tf.CategoryExpander
	}
	return f.Factory.CategoryExpander()
}

func (f *fakeAPIFactory) ClientSetForVersion(requiredVersion *schema.GroupVersion) (internalclientset.Interface, error) {
	return f.ClientSet()
}

func (f *fakeAPIFactory) ClientConfig() (*restclient.Config, error) {
	return f.tf.ClientConfig, f.tf.Err
}

func (f *fakeAPIFactory) ClientForMapping(m *meta.RESTMapping) (resource.RESTClient, error) {
	if f.tf.ClientForMappingFunc != nil {
		return f.tf.ClientForMappingFunc(m)
	}
	return f.tf.Client, f.tf.Err
}

func (f *fakeAPIFactory) UnstructuredClientForMapping(m *meta.RESTMapping) (resource.RESTClient, error) {
	if f.tf.UnstructuredClientForMappingFunc != nil {
		return f.tf.UnstructuredClientForMappingFunc(m)
	}
	return f.tf.UnstructuredClient, f.tf.Err
}

func (f *fakeAPIFactory) Describer(*meta.RESTMapping) (printers.Describer, error) {
	return f.tf.Describer, f.tf.Err
}

func (f *fakeAPIFactory) LogsForObject(object, options runtime.Object, timeout time.Duration) (*restclient.Request, error) {
	c, err := f.ClientSet()
	if err != nil {
		panic(err)
	}

	switch t := object.(type) {
	case *api.Pod:
		opts, ok := options.(*api.PodLogOptions)
		if !ok {
			return nil, errors.New("provided options object is not a PodLogOptions")
		}
		return c.Core().Pods(f.tf.Namespace).GetLogs(t.Name, opts), nil
	default:
		return nil, fmt.Errorf("cannot get the logs from %T", object)
	}
}

func (f *fakeAPIFactory) AttachablePodForObject(object runtime.Object, timeout time.Duration) (*api.Pod, error) {
	switch t := object.(type) {
	case *api.Pod:
		return t, nil
	default:
		return nil, fmt.Errorf("cannot attach to %T: not implemented", object)
	}
}

func (f *fakeAPIFactory) ApproximatePodTemplateForObject(obj runtime.Object) (*api.PodTemplateSpec, error) {
	return f.Factory.ApproximatePodTemplateForObject(obj)
}

func (f *fakeAPIFactory) Validator(validate bool) (validation.Schema, error) {
	return f.tf.Validator, f.tf.Err
}

func (f *fakeAPIFactory) DefaultNamespace() (string, bool, error) {
	return f.tf.Namespace, false, f.tf.Err
}

func (f *fakeAPIFactory) Command(*cobra.Command, bool) string {
	return f.tf.Command
}

func (f *fakeAPIFactory) Generators(cmdName string) map[string]kubectl.Generator {
	return cmdutil.DefaultGenerators(cmdName)
}

func (f *fakeAPIFactory) NewBuilder() *resource.Builder {
	mapper, typer := f.Object()

	return resource.NewBuilder(
		&resource.Mapper{
			RESTMapper:   mapper,
			ObjectTyper:  typer,
			ClientMapper: resource.ClientMapperFunc(f.ClientForMapping),
			Decoder:      cmdutil.InternalVersionDecoder(),
		},
		&resource.Mapper{
			RESTMapper:   mapper,
			ObjectTyper:  typer,
			ClientMapper: resource.ClientMapperFunc(f.UnstructuredClientForMapping),
			Decoder:      unstructured.UnstructuredJSONScheme,
		},
		f.CategoryExpander(),
	)
}

func (f *fakeAPIFactory) SuggestedPodTemplateResources() []schema.GroupResource {
	return []schema.GroupResource{}
}

func (f *fakeAPIFactory) OpenAPISchema() (openapi.Resources, error) {
	if f.tf.OpenAPISchemaFunc != nil {
		return f.tf.OpenAPISchemaFunc()
	}
	return openapitesting.EmptyResources{}, nil
}

func NewAPIFactory() (cmdutil.Factory, *TestFactory) {
	t := &TestFactory{
		Validator: validation.NullSchema{},
	}
	rf := cmdutil.NewFactory(nil)
	return &fakeAPIFactory{
		Factory: rf,
		tf:      t,
	}, t
}

func testDynamicResources() []*discovery.APIGroupResources {
	return []*discovery.APIGroupResources{
		{
			Group: metav1.APIGroup{
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1": {
					{Name: "pods", Namespaced: true, Kind: "Pod"},
					{Name: "services", Namespaced: true, Kind: "Service"},
					{Name: "replicationcontrollers", Namespaced: true, Kind: "ReplicationController"},
					{Name: "componentstatuses", Namespaced: false, Kind: "ComponentStatus"},
					{Name: "nodes", Namespaced: false, Kind: "Node"},
					{Name: "secrets", Namespaced: true, Kind: "Secret"},
					{Name: "configmaps", Namespaced: true, Kind: "ConfigMap"},
					{Name: "namespacedtype", Namespaced: true, Kind: "NamespacedType"},
					{Name: "namespaces", Namespaced: false, Kind: "Namespace"},
					{Name: "resourcequotas", Namespaced: true, Kind: "ResourceQuota"},
				},
			},
		},
		{
			Group: metav1.APIGroup{
				Name: "extensions",
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1beta1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1beta1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1beta1": {
					{Name: "deployments", Namespaced: true, Kind: "Deployment"},
					{Name: "replicasets", Namespaced: true, Kind: "ReplicaSet"},
				},
			},
		},
		{
			Group: metav1.APIGroup{
				Name: "apps",
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1beta1"},
					{Version: "v1beta2"},
					{Version: "v1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1beta1": {
					{Name: "deployments", Namespaced: true, Kind: "Deployment"},
					{Name: "replicasets", Namespaced: true, Kind: "ReplicaSet"},
				},
				"v1beta2": {
					{Name: "deployments", Namespaced: true, Kind: "Deployment"},
				},
				"v1": {
					{Name: "deployments", Namespaced: true, Kind: "Deployment"},
					{Name: "replicasets", Namespaced: true, Kind: "ReplicaSet"},
				},
			},
		},
		{
			Group: metav1.APIGroup{
				Name: "autoscaling",
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1"},
					{Version: "v2beta1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v2beta1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1": {
					{Name: "horizontalpodautoscalers", Namespaced: true, Kind: "HorizontalPodAutoscaler"},
				},
				"v2beta1": {
					{Name: "horizontalpodautoscalers", Namespaced: true, Kind: "HorizontalPodAutoscaler"},
				},
			},
		},
		{
			Group: metav1.APIGroup{
				Name: "storage.k8s.io",
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1beta1"},
					{Version: "v0"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1beta1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1beta1": {
					{Name: "storageclasses", Namespaced: false, Kind: "StorageClass"},
				},
				// bogus version of a known group/version/resource to make sure kubectl falls back to generic object mode
				"v0": {
					{Name: "storageclasses", Namespaced: false, Kind: "StorageClass"},
				},
			},
		},
		{
			Group: metav1.APIGroup{
				Name: "rbac.authorization.k8s.io",
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1beta1"},
					{Version: "v1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1": {
					{Name: "clusterroles", Namespaced: false, Kind: "ClusterRole"},
				},
				"v1beta1": {
					{Name: "clusterrolebindings", Namespaced: false, Kind: "ClusterRoleBinding"},
				},
			},
		},
		{
			Group: metav1.APIGroup{
				Name: "company.com",
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1": {
					{Name: "bars", Namespaced: true, Kind: "Bar"},
				},
			},
		},
		{
			Group: metav1.APIGroup{
				Name: "unit-test.test.com",
				Versions: []metav1.GroupVersionForDiscovery{
					{GroupVersion: "unit-test.test.com/v1", Version: "v1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{
					GroupVersion: "unit-test.test.com/v1",
					Version:      "v1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1": {
					{Name: "widgets", Namespaced: true, Kind: "Widget"},
				},
			},
		},
		{
			Group: metav1.APIGroup{
				Name: "apitest",
				Versions: []metav1.GroupVersionForDiscovery{
					{GroupVersion: "apitest/unlikelyversion", Version: "unlikelyversion"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{
					GroupVersion: "apitest/unlikelyversion",
					Version:      "unlikelyversion"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"unlikelyversion": {
					{Name: "types", SingularName: "type", Namespaced: false, Kind: "Type"},
				},
			},
		},
	}
}
