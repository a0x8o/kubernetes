/*
Copyright 2014 The Kubernetes Authors.

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

package master

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/kubernetes/pkg/api/unversioned"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	"k8s.io/kubernetes/pkg/apis/apps"
	appsapiv1alpha1 "k8s.io/kubernetes/pkg/apis/apps/v1alpha1"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	autoscalingapiv1 "k8s.io/kubernetes/pkg/apis/autoscaling/v1"
	"k8s.io/kubernetes/pkg/apis/batch"
	batchapiv1 "k8s.io/kubernetes/pkg/apis/batch/v1"
	batchapiv2alpha1 "k8s.io/kubernetes/pkg/apis/batch/v2alpha1"
	"k8s.io/kubernetes/pkg/apis/certificates"
	"k8s.io/kubernetes/pkg/apis/extensions"
	extensionsapiv1beta1 "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	"k8s.io/kubernetes/pkg/apis/rbac"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/generated/openapi"
	"k8s.io/kubernetes/pkg/genericapiserver"
	"k8s.io/kubernetes/pkg/kubelet/client"
	ipallocator "k8s.io/kubernetes/pkg/registry/core/service/ipallocator"
	"k8s.io/kubernetes/pkg/registry/registrytest"
	"k8s.io/kubernetes/pkg/runtime"
	etcdtesting "k8s.io/kubernetes/pkg/storage/etcd/testing"
	utilnet "k8s.io/kubernetes/pkg/util/net"
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/version"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/stretchr/testify/assert"
)

// setUp is a convience function for setting up for (most) tests.
func setUp(t *testing.T) (*Master, *etcdtesting.EtcdTestServer, Config, *assert.Assertions) {
	server, storageConfig := etcdtesting.NewUnsecuredEtcd3TestClientServer(t)

	config := &Config{
		GenericConfig: &genericapiserver.Config{},
	}

	resourceEncoding := genericapiserver.NewDefaultResourceEncodingConfig()
	resourceEncoding.SetVersionEncoding(api.GroupName, registered.GroupOrDie(api.GroupName).GroupVersion, unversioned.GroupVersion{Group: api.GroupName, Version: runtime.APIVersionInternal})
	resourceEncoding.SetVersionEncoding(autoscaling.GroupName, *testapi.Autoscaling.GroupVersion(), unversioned.GroupVersion{Group: autoscaling.GroupName, Version: runtime.APIVersionInternal})
	resourceEncoding.SetVersionEncoding(batch.GroupName, *testapi.Batch.GroupVersion(), unversioned.GroupVersion{Group: batch.GroupName, Version: runtime.APIVersionInternal})
	resourceEncoding.SetVersionEncoding(apps.GroupName, *testapi.Apps.GroupVersion(), unversioned.GroupVersion{Group: apps.GroupName, Version: runtime.APIVersionInternal})
	resourceEncoding.SetVersionEncoding(extensions.GroupName, *testapi.Extensions.GroupVersion(), unversioned.GroupVersion{Group: extensions.GroupName, Version: runtime.APIVersionInternal})
	resourceEncoding.SetVersionEncoding(rbac.GroupName, *testapi.Rbac.GroupVersion(), unversioned.GroupVersion{Group: rbac.GroupName, Version: runtime.APIVersionInternal})
	resourceEncoding.SetVersionEncoding(certificates.GroupName, *testapi.Certificates.GroupVersion(), unversioned.GroupVersion{Group: certificates.GroupName, Version: runtime.APIVersionInternal})
	storageFactory := genericapiserver.NewDefaultStorageFactory(*storageConfig, testapi.StorageMediaType(), api.Codecs, resourceEncoding, DefaultAPIResourceConfigSource())

	config.StorageFactory = storageFactory
	config.GenericConfig.LoopbackClientConfig = &restclient.Config{APIPath: "/api", ContentConfig: restclient.ContentConfig{NegotiatedSerializer: api.Codecs}}
	config.GenericConfig.APIResourceConfigSource = DefaultAPIResourceConfigSource()
	config.GenericConfig.PublicAddress = net.ParseIP("192.168.10.4")
	config.GenericConfig.Serializer = api.Codecs
	config.KubeletClient = client.FakeKubeletClient{}
	config.GenericConfig.APIPrefix = "/api"
	config.GenericConfig.APIGroupPrefix = "/apis"
	config.GenericConfig.APIResourceConfigSource = DefaultAPIResourceConfigSource()
	config.GenericConfig.ProxyDialer = func(network, addr string) (net.Conn, error) { return nil, nil }
	config.GenericConfig.ProxyTLSClientConfig = &tls.Config{}
	config.GenericConfig.RequestContextMapper = api.NewRequestContextMapper()
	config.GenericConfig.EnableVersion = true
	config.GenericConfig.LoopbackClientConfig = &restclient.Config{APIPath: "/api", ContentConfig: restclient.ContentConfig{NegotiatedSerializer: api.Codecs}}
	config.EnableCoreControllers = false

	// TODO: this is kind of hacky.  The trouble is that the sync loop
	// runs in a go-routine and there is no way to validate in the test
	// that the sync routine has actually run.  The right answer here
	// is probably to add some sort of callback that we can register
	// to validate that it's actually been run, but for now we don't
	// run the sync routine and register types manually.
	config.disableThirdPartyControllerForTesting = true

	master, err := config.Complete().New()
	if err != nil {
		t.Fatal(err)
	}

	fakeNodeClient := fake.NewSimpleClientset(registrytest.MakeNodeList([]string{"node1", "node2"}, api.NodeResources{}))
	master.nodeClient = fakeNodeClient.Core().Nodes()

	return master, server, *config, assert.New(t)
}

func newMaster(t *testing.T) (*Master, *etcdtesting.EtcdTestServer, Config, *assert.Assertions) {
	_, etcdserver, config, assert := setUp(t)

	master, err := config.Complete().New()
	if err != nil {
		t.Fatalf("Error in bringing up the master: %v", err)
	}

	return master, etcdserver, config, assert
}

// limitedAPIResourceConfigSource only enables the core group, the extensions group, the batch group, and the autoscaling group.
func limitedAPIResourceConfigSource() *genericapiserver.ResourceConfig {
	ret := genericapiserver.NewResourceConfig()
	ret.EnableVersions(
		apiv1.SchemeGroupVersion,
		extensionsapiv1beta1.SchemeGroupVersion,
		batchapiv1.SchemeGroupVersion,
		batchapiv2alpha1.SchemeGroupVersion,
		appsapiv1alpha1.SchemeGroupVersion,
		autoscalingapiv1.SchemeGroupVersion,
	)
	return ret
}

// newLimitedMaster only enables the core group, the extensions group, the batch group, and the autoscaling group.
func newLimitedMaster(t *testing.T) (*Master, *etcdtesting.EtcdTestServer, Config, *assert.Assertions) {
	_, etcdserver, config, assert := setUp(t)
	config.GenericConfig.APIResourceConfigSource = limitedAPIResourceConfigSource()
	master, err := config.Complete().New()
	if err != nil {
		t.Fatalf("Error in bringing up the master: %v", err)
	}

	return master, etcdserver, config, assert
}

// TestNew verifies that the New function returns a Master
// using the configuration properly.
func TestNew(t *testing.T) {
	master, etcdserver, config, assert := newMaster(t)
	defer etcdserver.Terminate(t)

	// Verify many of the variables match their config counterparts
	assert.Equal(master.RequestContextMapper(), config.GenericConfig.RequestContextMapper)
	assert.Equal(master.ClusterIP, config.GenericConfig.PublicAddress)

	// these values get defaulted
	_, serviceClusterIPRange, _ := net.ParseCIDR("10.0.0.0/24")
	serviceReadWriteIP, _ := ipallocator.GetIndexedIP(serviceClusterIPRange, 1)
	assert.Equal(master.MasterCount, 1)
	assert.Equal(master.PublicReadWritePort, 6443)
	assert.Equal(master.ServiceReadWriteIP, serviceReadWriteIP)

	// These functions should point to the same memory location
	masterDialer, _ := utilnet.Dialer(master.ProxyTransport)
	masterDialerFunc := fmt.Sprintf("%p", masterDialer)
	configDialerFunc := fmt.Sprintf("%p", config.GenericConfig.ProxyDialer)
	assert.Equal(masterDialerFunc, configDialerFunc)

	assert.Equal(master.ProxyTransport.(*http.Transport).TLSClientConfig, config.GenericConfig.ProxyTLSClientConfig)
}

// TestVersion tests /version
func TestVersion(t *testing.T) {
	s, etcdserver, _, _ := newMaster(t)
	defer etcdserver.Terminate(t)

	req, _ := http.NewRequest("GET", "/version", nil)
	resp := httptest.NewRecorder()
	s.InsecureHandler.ServeHTTP(resp, req)
	if resp.Code != 200 {
		t.Fatalf("expected http 200, got: %d", resp.Code)
	}

	var info version.Info
	err := json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(version.Get(), info) {
		t.Errorf("Expected %#v, Got %#v", version.Get(), info)
	}
}

// TestGetServersToValidate verifies the unexported getServersToValidate function
func TestGetServersToValidate(t *testing.T) {
	_, etcdserver, config, assert := setUp(t)
	defer etcdserver.Terminate(t)

	servers := getServersToValidate(config.StorageFactory)

	// Expected servers to validate: scheduler, controller-manager and etcd.
	assert.Equal(3, len(servers), "unexpected server list: %#v", servers)

	for _, server := range []string{"scheduler", "controller-manager", "etcd-0"} {
		if _, ok := servers[server]; !ok {
			t.Errorf("server list missing: %s", server)
		}
	}
}

// TestFindExternalAddress verifies both pass and fail cases for the unexported
// findExternalAddress function
func TestFindExternalAddress(t *testing.T) {
	assert := assert.New(t)
	expectedIP := "172.0.0.1"

	nodes := []*api.Node{new(api.Node), new(api.Node), new(api.Node)}
	nodes[0].Status.Addresses = []api.NodeAddress{{Type: "ExternalIP", Address: expectedIP}}
	nodes[1].Status.Addresses = []api.NodeAddress{{Type: "LegacyHostIP", Address: expectedIP}}
	nodes[2].Status.Addresses = []api.NodeAddress{{Type: "ExternalIP", Address: expectedIP}, {Type: "LegacyHostIP", Address: "172.0.0.2"}}

	// Pass Case
	for _, node := range nodes {
		ip, err := findExternalAddress(node)
		assert.NoError(err, "error getting node external address")
		assert.Equal(expectedIP, ip, "expected ip to be %s, but was %s", expectedIP, ip)
	}

	// Fail case
	_, err := findExternalAddress(new(api.Node))
	assert.Error(err, "expected findExternalAddress to fail on a node with missing ip information")
}

type fakeEndpointReconciler struct{}

func (*fakeEndpointReconciler) ReconcileEndpoints(serviceName string, ip net.IP, endpointPorts []api.EndpointPort, reconcilePorts bool) error {
	return nil
}

// TestGetNodeAddresses verifies that proper results are returned
// when requesting node addresses.
func TestGetNodeAddresses(t *testing.T) {
	master, etcdserver, _, assert := setUp(t)
	defer etcdserver.Terminate(t)

	// Fail case (no addresses associated with nodes)
	nodes, _ := master.nodeClient.List(api.ListOptions{})
	addrs, err := master.getNodeAddresses()

	assert.Error(err, "getNodeAddresses should have caused an error as there are no addresses.")
	assert.Equal([]string(nil), addrs)

	// Pass case with External type IP
	nodes, _ = master.nodeClient.List(api.ListOptions{})
	for index := range nodes.Items {
		nodes.Items[index].Status.Addresses = []api.NodeAddress{{Type: api.NodeExternalIP, Address: "127.0.0.1"}}
		master.nodeClient.Update(&nodes.Items[index])
	}
	addrs, err = master.getNodeAddresses()
	assert.NoError(err, "getNodeAddresses should not have returned an error.")
	assert.Equal([]string{"127.0.0.1", "127.0.0.1"}, addrs)

	// Pass case with LegacyHost type IP
	nodes, _ = master.nodeClient.List(api.ListOptions{})
	for index := range nodes.Items {
		nodes.Items[index].Status.Addresses = []api.NodeAddress{{Type: api.NodeLegacyHostIP, Address: "127.0.0.2"}}
		master.nodeClient.Update(&nodes.Items[index])
	}
	addrs, err = master.getNodeAddresses()
	assert.NoError(err, "getNodeAddresses failback should not have returned an error.")
	assert.Equal([]string{"127.0.0.2", "127.0.0.2"}, addrs)
}

func decodeResponse(resp *http.Response, obj interface{}) error {
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}
	return nil
}

// Because we need to be backwards compatible with release 1.1, at endpoints
// that exist in release 1.1, the responses should have empty APIVersion.
func TestAPIVersionOfDiscoveryEndpoints(t *testing.T) {
	master, etcdserver, _, assert := newMaster(t)
	defer etcdserver.Terminate(t)

	server := httptest.NewServer(master.HandlerContainer.ServeMux)

	// /api exists in release-1.1
	resp, err := http.Get(server.URL + "/api")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	apiVersions := unversioned.APIVersions{}
	assert.NoError(decodeResponse(resp, &apiVersions))
	assert.Equal(apiVersions.APIVersion, "")

	// /api/v1 exists in release-1.1
	resp, err = http.Get(server.URL + "/api/v1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	resourceList := unversioned.APIResourceList{}
	assert.NoError(decodeResponse(resp, &resourceList))
	assert.Equal(resourceList.APIVersion, "")

	// /apis exists in release-1.1
	resp, err = http.Get(server.URL + "/apis")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	groupList := unversioned.APIGroupList{}
	assert.NoError(decodeResponse(resp, &groupList))
	assert.Equal(groupList.APIVersion, "")

	// /apis/extensions exists in release-1.1
	resp, err = http.Get(server.URL + "/apis/extensions")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	group := unversioned.APIGroup{}
	assert.NoError(decodeResponse(resp, &group))
	assert.Equal(group.APIVersion, "")

	// /apis/extensions/v1beta1 exists in release-1.1
	resp, err = http.Get(server.URL + "/apis/extensions/v1beta1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	resourceList = unversioned.APIResourceList{}
	assert.NoError(decodeResponse(resp, &resourceList))
	assert.Equal(resourceList.APIVersion, "")

	// /apis/autoscaling doesn't exist in release-1.1, so the APIVersion field
	// should be non-empty in the results returned by the server.
	resp, err = http.Get(server.URL + "/apis/autoscaling")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	group = unversioned.APIGroup{}
	assert.NoError(decodeResponse(resp, &group))
	assert.Equal(group.APIVersion, "v1")

	// apis/autoscaling/v1 doesn't exist in release-1.1, so the APIVersion field
	// should be non-empty in the results returned by the server.

	resp, err = http.Get(server.URL + "/apis/autoscaling/v1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	resourceList = unversioned.APIResourceList{}
	assert.NoError(decodeResponse(resp, &resourceList))
	assert.Equal(resourceList.APIVersion, "v1")

}

func TestDiscoveryAtAPIS(t *testing.T) {
	master, etcdserver, _, assert := newLimitedMaster(t)
	defer etcdserver.Terminate(t)

	server := httptest.NewServer(master.HandlerContainer.ServeMux)
	resp, err := http.Get(server.URL + "/apis")
	if !assert.NoError(err) {
		t.Errorf("unexpected error: %v", err)
	}

	assert.Equal(http.StatusOK, resp.StatusCode)

	groupList := unversioned.APIGroupList{}
	assert.NoError(decodeResponse(resp, &groupList))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectGroupNames := sets.NewString(autoscaling.GroupName, batch.GroupName, apps.GroupName, extensions.GroupName)
	expectVersions := map[string][]unversioned.GroupVersionForDiscovery{
		autoscaling.GroupName: {
			{
				GroupVersion: testapi.Autoscaling.GroupVersion().String(),
				Version:      testapi.Autoscaling.GroupVersion().Version,
			},
		},
		// batch is using its pkg/apis/batch/ types here since during installation
		// both versions get installed and testapi.go currently does not support
		// multi-versioned clients
		batch.GroupName: {
			{
				GroupVersion: batchapiv1.SchemeGroupVersion.String(),
				Version:      batchapiv1.SchemeGroupVersion.Version,
			},
			{
				GroupVersion: batchapiv2alpha1.SchemeGroupVersion.String(),
				Version:      batchapiv2alpha1.SchemeGroupVersion.Version,
			},
		},
		apps.GroupName: {
			{
				GroupVersion: testapi.Apps.GroupVersion().String(),
				Version:      testapi.Apps.GroupVersion().Version,
			},
		},
		extensions.GroupName: {
			{
				GroupVersion: testapi.Extensions.GroupVersion().String(),
				Version:      testapi.Extensions.GroupVersion().Version,
			},
		},
	}
	expectPreferredVersion := map[string]unversioned.GroupVersionForDiscovery{
		autoscaling.GroupName: {
			GroupVersion: registered.GroupOrDie(autoscaling.GroupName).GroupVersion.String(),
			Version:      registered.GroupOrDie(autoscaling.GroupName).GroupVersion.Version,
		},
		batch.GroupName: {
			GroupVersion: registered.GroupOrDie(batch.GroupName).GroupVersion.String(),
			Version:      registered.GroupOrDie(batch.GroupName).GroupVersion.Version,
		},
		apps.GroupName: {
			GroupVersion: registered.GroupOrDie(apps.GroupName).GroupVersion.String(),
			Version:      registered.GroupOrDie(apps.GroupName).GroupVersion.Version,
		},
		extensions.GroupName: {
			GroupVersion: registered.GroupOrDie(extensions.GroupName).GroupVersion.String(),
			Version:      registered.GroupOrDie(extensions.GroupName).GroupVersion.Version,
		},
	}

	assert.Equal(4, len(groupList.Groups))
	for _, group := range groupList.Groups {
		if !expectGroupNames.Has(group.Name) {
			t.Errorf("got unexpected group %s", group.Name)
		}
		assert.Equal(expectVersions[group.Name], group.Versions)
		assert.Equal(expectPreferredVersion[group.Name], group.PreferredVersion)
	}

	thirdPartyGV := unversioned.GroupVersionForDiscovery{GroupVersion: "company.com/v1", Version: "v1"}
	master.addThirdPartyResourceStorage("/apis/company.com/v1", "foos", nil,
		unversioned.APIGroup{
			Name:             "company.com",
			Versions:         []unversioned.GroupVersionForDiscovery{thirdPartyGV},
			PreferredVersion: thirdPartyGV,
		})

	resp, err = http.Get(server.URL + "/apis")
	if !assert.NoError(err) {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.NoError(decodeResponse(resp, &groupList))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Equal(5, len(groupList.Groups))

	expectGroupNames.Insert("company.com")
	expectVersions["company.com"] = []unversioned.GroupVersionForDiscovery{thirdPartyGV}
	expectPreferredVersion["company.com"] = thirdPartyGV
	for _, group := range groupList.Groups {
		if !expectGroupNames.Has(group.Name) {
			t.Errorf("got unexpected group %s", group.Name)
		}
		assert.Equal(expectVersions[group.Name], group.Versions)
		assert.Equal(expectPreferredVersion[group.Name], group.PreferredVersion)
	}
}

func writeResponseToFile(resp *http.Response, filename string) error {
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0755)
}

// TestValidOpenAPISpec verifies that the open api is added
// at the proper endpoint and the spec is valid.
func TestValidOpenAPISpec(t *testing.T) {
	_, etcdserver, config, assert := setUp(t)
	defer etcdserver.Terminate(t)

	config.GenericConfig.OpenAPIDefinitions = openapi.OpenAPIDefinitions
	config.GenericConfig.EnableOpenAPISupport = true
	config.GenericConfig.EnableIndex = true
	config.GenericConfig.OpenAPIInfo = spec.Info{
		InfoProps: spec.InfoProps{
			Title:   "Kubernetes",
			Version: "unversioned",
		},
	}
	master, err := config.Complete().New()
	if err != nil {
		t.Fatalf("Error in bringing up the master: %v", err)
	}

	// make sure swagger.json is not registered before calling install api.
	server := httptest.NewServer(master.HandlerContainer.ServeMux)
	resp, err := http.Get(server.URL + "/swagger.json")
	if !assert.NoError(err) {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(http.StatusNotFound, resp.StatusCode)

	master.InstallOpenAPI()
	resp, err = http.Get(server.URL + "/swagger.json")
	if !assert.NoError(err) {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(http.StatusOK, resp.StatusCode)

	// as json schema
	var sch spec.Schema
	if assert.NoError(decodeResponse(resp, &sch)) {
		validator := validate.NewSchemaValidator(spec.MustLoadSwagger20Schema(), nil, "", strfmt.Default)
		res := validator.Validate(&sch)
		assert.NoError(res.AsError())
	}

	// TODO(mehdy): The actual validation part of these tests are timing out on jerkin but passing locally. Enable it after debugging timeout issue.
	disableValidation := true

	// Saving specs to a temporary folder is a good way to debug spec generation without bringing up an actual
	// api server.
	saveSwaggerSpecs := false

	// Validate OpenApi spec
	doc, err := loads.Spec(server.URL + "/swagger.json")
	if assert.NoError(err) {
		validator := validate.NewSpecValidator(doc.Schema(), strfmt.Default)
		if !disableValidation {
			res, warns := validator.Validate(doc)
			assert.NoError(res.AsError())
			if !warns.IsValid() {
				t.Logf("Open API spec on root has some warnings : %v", warns)
			}
		} else {
			t.Logf("Validation is disabled because it is timing out on jenkins put passing locally.")
		}
	}

	// validate specs on each end-point
	resp, err = http.Get(server.URL)
	if !assert.NoError(err) {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(http.StatusOK, resp.StatusCode)
	var list unversioned.RootPaths
	if assert.NoError(decodeResponse(resp, &list)) {
		for _, path := range list.Paths {
			if !strings.HasPrefix(path, "/api") {
				continue
			}
			t.Logf("Validating open API spec on %v ...", path)

			if saveSwaggerSpecs {
				resp, err = http.Get(server.URL + path + "/swagger.json")
				if !assert.NoError(err) {
					t.Errorf("unexpected error: %v", err)
				}
				assert.Equal(http.StatusOK, resp.StatusCode)
				assert.NoError(writeResponseToFile(resp, "/tmp/swagger_"+strings.Replace(path, "/", "_", -1)+".json"))
			}

			// Validate OpenApi spec on path
			doc, err := loads.Spec(server.URL + path + "/swagger.json")
			if assert.NoError(err) {
				validator := validate.NewSpecValidator(doc.Schema(), strfmt.Default)
				if !disableValidation {
					res, warns := validator.Validate(doc)
					assert.NoError(res.AsError())
					if !warns.IsValid() {
						t.Logf("Open API spec on %v has some warnings : %v", path, warns)
					}
				} else {
					t.Logf("Validation is disabled because it is timing out on jenkins put passing locally.")
				}
			}

		}
	}

}
