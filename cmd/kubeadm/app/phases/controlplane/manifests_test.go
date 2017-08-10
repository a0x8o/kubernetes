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

package controlplane

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/yaml"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/pkg/util/version"
)

const (
	testCertsDir = "/var/lib/certs"
	etcdDataDir  = "/var/lib/etcd"
)

func TestWriteStaticPodManifests(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	// set up tmp KubernetesDir for testing
	kubeadmconstants.KubernetesDir = fmt.Sprintf("%s/etc/kubernetes", tmpdir)
	defer func() { kubeadmconstants.KubernetesDir = "/etc/kubernetes" }()

	var tests = []struct {
		cfg                  *kubeadmapi.MasterConfiguration
		expectErr            bool
		expectedAPIProbePort int32
	}{
		{
			cfg: &kubeadmapi.MasterConfiguration{
				KubernetesVersion: "v1.7.0",
			},
			expectErr: false,
		},
		{
			cfg: &kubeadmapi.MasterConfiguration{
				API: kubeadmapi.API{
					BindPort: 443,
				},
				KubernetesVersion: "v1.7.0",
			},
			expectErr:            false,
			expectedAPIProbePort: 443,
		},
	}
	for _, rt := range tests {

		actual := WriteStaticPodManifests(rt.cfg, version.MustParseSemantic(rt.cfg.KubernetesVersion), fmt.Sprintf("%s/etc/kubernetes/manifests", tmpdir))
		if (actual == nil) && rt.expectErr {
			t.Error("expected an error from WriteStaticPodManifests but got none")
			continue
		}
		if (actual != nil) && !rt.expectErr {
			t.Errorf("didn't expect an error from WriteStaticPodManifests but got: %v", err)
			continue
		}
		if rt.expectErr {
			continue
		}

		// Below is dead code.
		if rt.expectedAPIProbePort != 0 {
			manifest, err := os.Open(filepath.Join(kubeadmconstants.KubernetesDir, kubeadmconstants.ManifestsSubDirName, "kube-apiserver.yaml"))
			if err != nil {
				t.Errorf("WriteStaticPodManifests: %v", err)
				continue
			}
			defer manifest.Close()

			var pod v1.Pod
			d := yaml.NewYAMLOrJSONDecoder(manifest, 4096)
			if err := d.Decode(&pod); err != nil {
				t.Error("WriteStaticPodManifests: error decoding manifests/kube-apiserver.yaml into Pod")
				continue
			}

			// Lots of individual checks as we traverse pointers so we don't panic dereferencing a nil on failure
			containers := pod.Spec.Containers
			if containers == nil || len(containers) == 0 {
				t.Error("WriteStaticPodManifests: wrote an apiserver manifest without any containers")
				continue
			}

			probe := containers[0].LivenessProbe
			if probe == nil {
				t.Error("WriteStaticPodManifests: wrote an apiserver manifest without a liveness probe")
				continue
			}

			httpGET := probe.Handler.HTTPGet
			if httpGET == nil {
				t.Error("WriteStaticPodManifests: wrote an apiserver manifest without an HTTP liveness probe")
				continue
			}

			port := httpGET.Port.IntVal
			if rt.expectedAPIProbePort != port {
				t.Errorf("WriteStaticPodManifests: apiserver pod liveness probe port was: %v, wanted %v", port, rt.expectedAPIProbePort)
			}
		}
	}
}

func TestComponentResources(t *testing.T) {
	a := componentResources("250m")
	if a.Requests == nil {
		t.Errorf(
			"failed componentResources, return value was nil",
		)
	}
}

func TestComponentProbe(t *testing.T) {
	var tests = []struct {
		port   int
		path   string
		scheme v1.URIScheme
	}{
		{
			port:   1,
			path:   "foo",
			scheme: v1.URISchemeHTTP,
		},
		{
			port:   2,
			path:   "bar",
			scheme: v1.URISchemeHTTPS,
		},
	}
	for _, rt := range tests {
		actual := componentProbe(rt.port, rt.path, rt.scheme)
		if actual.Handler.HTTPGet.Port != intstr.FromInt(rt.port) {
			t.Errorf(
				"failed componentProbe:\n\texpected: %v\n\t  actual: %v",
				rt.port,
				actual.Handler.HTTPGet.Port,
			)
		}
		if actual.Handler.HTTPGet.Path != rt.path {
			t.Errorf(
				"failed componentProbe:\n\texpected: %s\n\t  actual: %s",
				rt.path,
				actual.Handler.HTTPGet.Path,
			)
		}
		if actual.Handler.HTTPGet.Scheme != rt.scheme {
			t.Errorf(
				"failed componentProbe:\n\texpected: %v\n\t  actual: %v",
				rt.scheme,
				actual.Handler.HTTPGet.Scheme,
			)
		}
	}
}

func TestComponentPod(t *testing.T) {
	var tests = []struct {
		name     string
		expected v1.Pod
	}{
		{
			name: "foo",
			expected: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "foo",
					Namespace:   "kube-system",
					Annotations: map[string]string{"scheduler.alpha.kubernetes.io/critical-pod": ""},
					Labels:      map[string]string{"component": "foo", "tier": "control-plane"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "foo",
						},
					},
					HostNetwork: true,
					Volumes:     []v1.Volume{},
				},
			},
		},
	}

	for _, rt := range tests {
		c := v1.Container{Name: rt.name}
		actual := componentPod(c, []v1.Volume{})
		if !reflect.DeepEqual(rt.expected, actual) {
			t.Errorf(
				"failed componentPod:\n\texpected: %v\n\t  actual: %v",
				rt.expected,
				actual,
			)
		}
	}
}

func TestGetAPIServerCommand(t *testing.T) {
	var tests = []struct {
		cfg      *kubeadmapi.MasterConfiguration
		expected []string
	}{
		{
			cfg: &kubeadmapi.MasterConfiguration{
				API:               kubeadmapi.API{BindPort: 123, AdvertiseAddress: "1.2.3.4"},
				Networking:        kubeadmapi.Networking{ServiceSubnet: "bar"},
				CertificatesDir:   testCertsDir,
				KubernetesVersion: "v1.7.0",
			},
			expected: []string{
				"kube-apiserver",
				"--insecure-port=0",
				"--admission-control=Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,ResourceQuota",
				"--service-cluster-ip-range=bar",
				"--service-account-key-file=" + testCertsDir + "/sa.pub",
				"--client-ca-file=" + testCertsDir + "/ca.crt",
				"--tls-cert-file=" + testCertsDir + "/apiserver.crt",
				"--tls-private-key-file=" + testCertsDir + "/apiserver.key",
				"--kubelet-client-certificate=" + testCertsDir + "/apiserver-kubelet-client.crt",
				"--kubelet-client-key=" + testCertsDir + "/apiserver-kubelet-client.key",
				"--secure-port=123",
				"--allow-privileged=true",
				"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
				"--experimental-bootstrap-token-auth=true",
				"--proxy-client-cert-file=/var/lib/certs/front-proxy-client.crt",
				"--proxy-client-key-file=/var/lib/certs/front-proxy-client.key",
				"--requestheader-username-headers=X-Remote-User",
				"--requestheader-group-headers=X-Remote-Group",
				"--requestheader-extra-headers-prefix=X-Remote-Extra-",
				"--requestheader-client-ca-file=" + testCertsDir + "/front-proxy-ca.crt",
				"--requestheader-allowed-names=front-proxy-client",
				"--authorization-mode=Node,RBAC",
				"--advertise-address=1.2.3.4",
				"--etcd-servers=http://127.0.0.1:2379",
			},
		},
		{
			cfg: &kubeadmapi.MasterConfiguration{
				API:               kubeadmapi.API{BindPort: 123, AdvertiseAddress: "4.3.2.1"},
				Networking:        kubeadmapi.Networking{ServiceSubnet: "bar"},
				CertificatesDir:   testCertsDir,
				KubernetesVersion: "v1.7.1",
			},
			expected: []string{
				"kube-apiserver",
				"--insecure-port=0",
				"--admission-control=Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,ResourceQuota",
				"--service-cluster-ip-range=bar",
				"--service-account-key-file=" + testCertsDir + "/sa.pub",
				"--client-ca-file=" + testCertsDir + "/ca.crt",
				"--tls-cert-file=" + testCertsDir + "/apiserver.crt",
				"--tls-private-key-file=" + testCertsDir + "/apiserver.key",
				"--kubelet-client-certificate=" + testCertsDir + "/apiserver-kubelet-client.crt",
				"--kubelet-client-key=" + testCertsDir + "/apiserver-kubelet-client.key",
				"--secure-port=123",
				"--allow-privileged=true",
				"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
				"--experimental-bootstrap-token-auth=true",
				"--proxy-client-cert-file=/var/lib/certs/front-proxy-client.crt",
				"--proxy-client-key-file=/var/lib/certs/front-proxy-client.key",
				"--requestheader-username-headers=X-Remote-User",
				"--requestheader-group-headers=X-Remote-Group",
				"--requestheader-extra-headers-prefix=X-Remote-Extra-",
				"--requestheader-client-ca-file=" + testCertsDir + "/front-proxy-ca.crt",
				"--requestheader-allowed-names=front-proxy-client",
				"--authorization-mode=Node,RBAC",
				"--advertise-address=4.3.2.1",
				"--etcd-servers=http://127.0.0.1:2379",
			},
		},
		{
			cfg: &kubeadmapi.MasterConfiguration{
				API:               kubeadmapi.API{BindPort: 123, AdvertiseAddress: "4.3.2.1"},
				Networking:        kubeadmapi.Networking{ServiceSubnet: "bar"},
				Etcd:              kubeadmapi.Etcd{CertFile: "fiz", KeyFile: "faz"},
				CertificatesDir:   testCertsDir,
				KubernetesVersion: "v1.7.2",
			},
			expected: []string{
				"kube-apiserver",
				"--insecure-port=0",
				"--admission-control=Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,ResourceQuota",
				"--service-cluster-ip-range=bar",
				"--service-account-key-file=" + testCertsDir + "/sa.pub",
				"--client-ca-file=" + testCertsDir + "/ca.crt",
				"--tls-cert-file=" + testCertsDir + "/apiserver.crt",
				"--tls-private-key-file=" + testCertsDir + "/apiserver.key",
				"--kubelet-client-certificate=" + testCertsDir + "/apiserver-kubelet-client.crt",
				"--kubelet-client-key=" + testCertsDir + "/apiserver-kubelet-client.key",
				"--secure-port=123",
				"--allow-privileged=true",
				"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
				"--experimental-bootstrap-token-auth=true",
				"--proxy-client-cert-file=/var/lib/certs/front-proxy-client.crt",
				"--proxy-client-key-file=/var/lib/certs/front-proxy-client.key",
				"--requestheader-username-headers=X-Remote-User",
				"--requestheader-group-headers=X-Remote-Group",
				"--requestheader-extra-headers-prefix=X-Remote-Extra-",
				"--requestheader-client-ca-file=" + testCertsDir + "/front-proxy-ca.crt",
				"--requestheader-allowed-names=front-proxy-client",
				"--authorization-mode=Node,RBAC",
				"--advertise-address=4.3.2.1",
				"--etcd-servers=http://127.0.0.1:2379",
				"--etcd-certfile=fiz",
				"--etcd-keyfile=faz",
			},
		},
		{
			cfg: &kubeadmapi.MasterConfiguration{
				API:               kubeadmapi.API{BindPort: 123, AdvertiseAddress: "4.3.2.1"},
				Networking:        kubeadmapi.Networking{ServiceSubnet: "bar"},
				Etcd:              kubeadmapi.Etcd{CertFile: "fiz", KeyFile: "faz"},
				CertificatesDir:   testCertsDir,
				KubernetesVersion: "v1.7.3",
			},
			expected: []string{
				"kube-apiserver",
				"--insecure-port=0",
				"--admission-control=Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,ResourceQuota",
				"--service-cluster-ip-range=bar",
				"--service-account-key-file=" + testCertsDir + "/sa.pub",
				"--client-ca-file=" + testCertsDir + "/ca.crt",
				"--tls-cert-file=" + testCertsDir + "/apiserver.crt",
				"--tls-private-key-file=" + testCertsDir + "/apiserver.key",
				"--kubelet-client-certificate=" + testCertsDir + "/apiserver-kubelet-client.crt",
				"--kubelet-client-key=" + testCertsDir + "/apiserver-kubelet-client.key",
				fmt.Sprintf("--secure-port=%d", 123),
				"--allow-privileged=true",
				"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
				"--experimental-bootstrap-token-auth=true",
				"--proxy-client-cert-file=/var/lib/certs/front-proxy-client.crt",
				"--proxy-client-key-file=/var/lib/certs/front-proxy-client.key",
				"--requestheader-username-headers=X-Remote-User",
				"--requestheader-group-headers=X-Remote-Group",
				"--requestheader-extra-headers-prefix=X-Remote-Extra-",
				"--requestheader-client-ca-file=" + testCertsDir + "/front-proxy-ca.crt",
				"--requestheader-allowed-names=front-proxy-client",
				"--authorization-mode=Node,RBAC",
				"--advertise-address=4.3.2.1",
				"--etcd-servers=http://127.0.0.1:2379",
				"--etcd-certfile=fiz",
				"--etcd-keyfile=faz",
			},
		},
		{
			cfg: &kubeadmapi.MasterConfiguration{
				API:               kubeadmapi.API{BindPort: 123, AdvertiseAddress: "2001:db8::1"},
				Networking:        kubeadmapi.Networking{ServiceSubnet: "bar"},
				Etcd:              kubeadmapi.Etcd{CertFile: "fiz", KeyFile: "faz"},
				CertificatesDir:   testCertsDir,
				KubernetesVersion: "v1.7.0",
			},
			expected: []string{
				"kube-apiserver",
				"--insecure-port=0",
				"--admission-control=Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,ResourceQuota",
				"--service-cluster-ip-range=bar",
				"--service-account-key-file=" + testCertsDir + "/sa.pub",
				"--client-ca-file=" + testCertsDir + "/ca.crt",
				"--tls-cert-file=" + testCertsDir + "/apiserver.crt",
				"--tls-private-key-file=" + testCertsDir + "/apiserver.key",
				"--kubelet-client-certificate=" + testCertsDir + "/apiserver-kubelet-client.crt",
				"--kubelet-client-key=" + testCertsDir + "/apiserver-kubelet-client.key",
				fmt.Sprintf("--secure-port=%d", 123),
				"--allow-privileged=true",
				"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
				"--experimental-bootstrap-token-auth=true",
				"--proxy-client-cert-file=/var/lib/certs/front-proxy-client.crt",
				"--proxy-client-key-file=/var/lib/certs/front-proxy-client.key",
				"--requestheader-username-headers=X-Remote-User",
				"--requestheader-group-headers=X-Remote-Group",
				"--requestheader-extra-headers-prefix=X-Remote-Extra-",
				"--requestheader-client-ca-file=" + testCertsDir + "/front-proxy-ca.crt",
				"--requestheader-allowed-names=front-proxy-client",
				"--authorization-mode=Node,RBAC",
				"--advertise-address=2001:db8::1",
				"--etcd-servers=http://127.0.0.1:2379",
				"--etcd-certfile=fiz",
				"--etcd-keyfile=faz",
			},
		},
	}

	for _, rt := range tests {
		actual := getAPIServerCommand(rt.cfg, version.MustParseSemantic(rt.cfg.KubernetesVersion))
		sort.Strings(actual)
		sort.Strings(rt.expected)
		if !reflect.DeepEqual(actual, rt.expected) {
			t.Errorf("failed getAPIServerCommand:\nexpected:\n%v\nsaw:\n%v", rt.expected, actual)
		}
	}
}

func TestGetControllerManagerCommand(t *testing.T) {
	var tests = []struct {
		cfg      *kubeadmapi.MasterConfiguration
		expected []string
	}{
		{
			cfg: &kubeadmapi.MasterConfiguration{
				CertificatesDir:   testCertsDir,
				KubernetesVersion: "v1.7.0",
			},
			expected: []string{
				"kube-controller-manager",
				"--address=127.0.0.1",
				"--leader-elect=true",
				"--kubeconfig=" + kubeadmconstants.KubernetesDir + "/controller-manager.conf",
				"--root-ca-file=" + testCertsDir + "/ca.crt",
				"--service-account-private-key-file=" + testCertsDir + "/sa.key",
				"--cluster-signing-cert-file=" + testCertsDir + "/ca.crt",
				"--cluster-signing-key-file=" + testCertsDir + "/ca.key",
				"--use-service-account-credentials=true",
				"--controllers=*,bootstrapsigner,tokencleaner",
			},
		},
		{
			cfg: &kubeadmapi.MasterConfiguration{
				CloudProvider:     "foo",
				CertificatesDir:   testCertsDir,
				KubernetesVersion: "v1.7.0",
			},
			expected: []string{
				"kube-controller-manager",
				"--address=127.0.0.1",
				"--leader-elect=true",
				"--kubeconfig=" + kubeadmconstants.KubernetesDir + "/controller-manager.conf",
				"--root-ca-file=" + testCertsDir + "/ca.crt",
				"--service-account-private-key-file=" + testCertsDir + "/sa.key",
				"--cluster-signing-cert-file=" + testCertsDir + "/ca.crt",
				"--cluster-signing-key-file=" + testCertsDir + "/ca.key",
				"--use-service-account-credentials=true",
				"--controllers=*,bootstrapsigner,tokencleaner",
				"--cloud-provider=foo",
			},
		},
		{
			cfg: &kubeadmapi.MasterConfiguration{
				Networking:        kubeadmapi.Networking{PodSubnet: "bar"},
				CertificatesDir:   testCertsDir,
				KubernetesVersion: "v1.7.0",
			},
			expected: []string{
				"kube-controller-manager",
				"--address=127.0.0.1",
				"--leader-elect=true",
				"--kubeconfig=" + kubeadmconstants.KubernetesDir + "/controller-manager.conf",
				"--root-ca-file=" + testCertsDir + "/ca.crt",
				"--service-account-private-key-file=" + testCertsDir + "/sa.key",
				"--cluster-signing-cert-file=" + testCertsDir + "/ca.crt",
				"--cluster-signing-key-file=" + testCertsDir + "/ca.key",
				"--use-service-account-credentials=true",
				"--controllers=*,bootstrapsigner,tokencleaner",
				"--allocate-node-cidrs=true",
				"--cluster-cidr=bar",
			},
		},
	}

	for _, rt := range tests {
		actual := getControllerManagerCommand(rt.cfg, version.MustParseSemantic(rt.cfg.KubernetesVersion))
		sort.Strings(actual)
		sort.Strings(rt.expected)
		if !reflect.DeepEqual(actual, rt.expected) {
			t.Errorf("failed getControllerManagerCommand:\nexpected:\n%v\nsaw:\n%v", rt.expected, actual)
		}
	}
}

func TestGetEtcdCommand(t *testing.T) {
	var tests = []struct {
		cfg      *kubeadmapi.MasterConfiguration
		expected []string
	}{
		{
			cfg: &kubeadmapi.MasterConfiguration{
				Etcd: kubeadmapi.Etcd{DataDir: "/var/lib/etcd"},
			},
			expected: []string{
				"etcd",
				"--listen-client-urls=http://127.0.0.1:2379",
				"--advertise-client-urls=http://127.0.0.1:2379",
				"--data-dir=/var/lib/etcd",
			},
		},
		{
			cfg: &kubeadmapi.MasterConfiguration{
				Etcd: kubeadmapi.Etcd{
					DataDir: "/var/lib/etcd",
					ExtraArgs: map[string]string{
						"listen-client-urls":    "http://10.0.1.10:2379",
						"advertise-client-urls": "http://10.0.1.10:2379",
					},
				},
			},
			expected: []string{
				"etcd",
				"--listen-client-urls=http://10.0.1.10:2379",
				"--advertise-client-urls=http://10.0.1.10:2379",
				"--data-dir=/var/lib/etcd",
			},
		},
		{
			cfg: &kubeadmapi.MasterConfiguration{
				Etcd: kubeadmapi.Etcd{DataDir: "/etc/foo"},
			},
			expected: []string{
				"etcd",
				"--listen-client-urls=http://127.0.0.1:2379",
				"--advertise-client-urls=http://127.0.0.1:2379",
				"--data-dir=/etc/foo",
			},
		},
	}

	for _, rt := range tests {
		actual := getEtcdCommand(rt.cfg)
		sort.Strings(actual)
		sort.Strings(rt.expected)
		if !reflect.DeepEqual(actual, rt.expected) {
			t.Errorf("failed getEtcdCommand:\nexpected:\n%v\nsaw:\n%v", rt.expected, actual)
		}
	}
}

func TestGetSchedulerCommand(t *testing.T) {
	var tests = []struct {
		cfg      *kubeadmapi.MasterConfiguration
		expected []string
	}{
		{
			cfg: &kubeadmapi.MasterConfiguration{},
			expected: []string{
				"kube-scheduler",
				"--address=127.0.0.1",
				"--leader-elect=true",
				"--kubeconfig=" + kubeadmconstants.KubernetesDir + "/scheduler.conf",
			},
		},
	}

	for _, rt := range tests {
		actual := getSchedulerCommand(rt.cfg)
		sort.Strings(actual)
		sort.Strings(rt.expected)
		if !reflect.DeepEqual(actual, rt.expected) {
			t.Errorf("failed getSchedulerCommand:\nexpected:\n%v\nsaw:\n%v", rt.expected, actual)
		}
	}
}

func TestGetAuthzParameters(t *testing.T) {
	var tests = []struct {
		authMode []string
		expected []string
	}{
		{
			authMode: []string{},
			expected: []string{
				"--authorization-mode=Node,RBAC",
			},
		},
		{
			authMode: []string{"RBAC"},
			expected: []string{
				"--authorization-mode=RBAC",
			},
		},
		{
			authMode: []string{"AlwaysAllow"},
			expected: []string{
				"--authorization-mode=AlwaysAllow",
			},
		},
		{
			authMode: []string{"AlwaysDeny"},
			expected: []string{
				"--authorization-mode=AlwaysDeny",
			},
		},
		{
			authMode: []string{"ABAC"},
			expected: []string{
				"--authorization-mode=ABAC",
				"--authorization-policy-file=/etc/kubernetes/abac_policy.json",
			},
		},
		{
			authMode: []string{"ABAC", "Webhook"},
			expected: []string{
				"--authorization-mode=ABAC,Webhook",
				"--authorization-policy-file=/etc/kubernetes/abac_policy.json",
				"--authorization-webhook-config-file=/etc/kubernetes/webhook_authz.conf",
			},
		},
		{
			authMode: []string{"ABAC", "RBAC", "Webhook"},
			expected: []string{
				"--authorization-mode=ABAC,RBAC,Webhook",
				"--authorization-policy-file=/etc/kubernetes/abac_policy.json",
				"--authorization-webhook-config-file=/etc/kubernetes/webhook_authz.conf",
			},
		},
		{
			authMode: []string{"Node", "RBAC", "Webhook", "ABAC"},
			expected: []string{
				"--authorization-mode=Node,RBAC,Webhook,ABAC",
				"--authorization-policy-file=/etc/kubernetes/abac_policy.json",
				"--authorization-webhook-config-file=/etc/kubernetes/webhook_authz.conf",
			},
		},
	}

	for _, rt := range tests {
		actual := getAuthzParameters(rt.authMode)
		sort.Strings(actual)
		sort.Strings(rt.expected)
		if !reflect.DeepEqual(actual, rt.expected) {
			t.Errorf("failed getAuthzParameters:\nexpected:\n%v\nsaw:\n%v", rt.expected, actual)
		}
	}
}

func TestGetExtraParameters(t *testing.T) {
	var tests = []struct {
		overrides map[string]string
		defaults  map[string]string
		expected  []string
	}{
		{
			overrides: map[string]string{
				"admission-control": "NamespaceLifecycle,LimitRanger",
			},
			defaults: map[string]string{
				"admission-control":     "NamespaceLifecycle",
				"insecure-bind-address": "127.0.0.1",
				"allow-privileged":      "true",
			},
			expected: []string{
				"--admission-control=NamespaceLifecycle,LimitRanger",
				"--insecure-bind-address=127.0.0.1",
				"--allow-privileged=true",
			},
		},
		{
			overrides: map[string]string{
				"admission-control": "NamespaceLifecycle,LimitRanger",
			},
			defaults: map[string]string{
				"insecure-bind-address": "127.0.0.1",
				"allow-privileged":      "true",
			},
			expected: []string{
				"--admission-control=NamespaceLifecycle,LimitRanger",
				"--insecure-bind-address=127.0.0.1",
				"--allow-privileged=true",
			},
		},
	}

	for _, rt := range tests {
		actual := getExtraParameters(rt.overrides, rt.defaults)
		sort.Strings(actual)
		sort.Strings(rt.expected)
		if !reflect.DeepEqual(actual, rt.expected) {
			t.Errorf("failed getExtraParameters:\nexpected:\n%v\nsaw:\n%v", rt.expected, actual)
		}
	}
}
