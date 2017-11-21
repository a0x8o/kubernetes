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

package v1alpha1

import (
	"net/url"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubeletconfigv1alpha1 "k8s.io/kubernetes/pkg/kubelet/apis/kubeletconfig/v1alpha1"
	utilpointer "k8s.io/kubernetes/pkg/util/pointer"
)

const (
	// DefaultServiceDNSDomain defines default cluster-internal domain name for Services and Pods
	DefaultServiceDNSDomain = "cluster.local"
	// DefaultServicesSubnet defines default service subnet range
	DefaultServicesSubnet = "10.96.0.0/12"
	// DefaultClusterDNSIP defines default DNS IP
	DefaultClusterDNSIP = "10.96.0.10"
	// DefaultKubernetesVersion defines default kubernetes version
	DefaultKubernetesVersion = "stable-1.8"
	// DefaultAPIBindPort defines default API port
	DefaultAPIBindPort = 6443
	// DefaultAuthorizationModes defines default authorization modes
	DefaultAuthorizationModes = "Node,RBAC"
	// DefaultCertificatesDir defines default certificate directory
	DefaultCertificatesDir = "/etc/kubernetes/pki"
	// DefaultImageRepository defines default image registry
	DefaultImageRepository = "gcr.io/google_containers"
	// DefaultManifestsDir defines default manifests directory
	DefaultManifestsDir = "/etc/kubernetes/manifests"

	// DefaultEtcdDataDir defines default location of etcd where static pods will save data to
	DefaultEtcdDataDir = "/var/lib/etcd"
	// DefaultEtcdClusterSize defines the default cluster size when using the etcd-operator
	DefaultEtcdClusterSize = 3
	// DefaultEtcdOperatorVersion defines the default version of the etcd-operator to use
	DefaultEtcdOperatorVersion = "v0.6.0"
	// DefaultEtcdCertDir represents the directory where PKI assets are stored for self-hosted etcd
	DefaultEtcdCertDir = "/etc/kubernetes/pki/etcd"
	// DefaultEtcdClusterServiceName is the default name of the service backing the etcd cluster
	DefaultEtcdClusterServiceName = "etcd-cluster"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_MasterConfiguration assigns default values to Master node
func SetDefaults_MasterConfiguration(obj *MasterConfiguration) {
	if obj.KubernetesVersion == "" {
		obj.KubernetesVersion = DefaultKubernetesVersion
	}

	if obj.API.BindPort == 0 {
		obj.API.BindPort = DefaultAPIBindPort
	}

	if obj.Networking.ServiceSubnet == "" {
		obj.Networking.ServiceSubnet = DefaultServicesSubnet
	}

	if obj.Networking.DNSDomain == "" {
		obj.Networking.DNSDomain = DefaultServiceDNSDomain
	}

	if len(obj.AuthorizationModes) == 0 {
		obj.AuthorizationModes = strings.Split(DefaultAuthorizationModes, ",")
	}

	if obj.CertificatesDir == "" {
		obj.CertificatesDir = DefaultCertificatesDir
	}

	if obj.TokenTTL == nil {
		obj.TokenTTL = &metav1.Duration{
			Duration: constants.DefaultTokenDuration,
		}
	}

	if obj.ImageRepository == "" {
		obj.ImageRepository = DefaultImageRepository
	}

	if obj.Etcd.DataDir == "" {
		obj.Etcd.DataDir = DefaultEtcdDataDir
	}

	SetDefaultsEtcdSelfHosted(obj)
	SetDefaults_KubeletConfiguration(obj)
}

// SetDefaults_NodeConfiguration assigns default values to a regular node
func SetDefaults_NodeConfiguration(obj *NodeConfiguration) {
	if obj.CACertPath == "" {
		obj.CACertPath = DefaultCACertPath
	}
	if len(obj.TLSBootstrapToken) == 0 {
		obj.TLSBootstrapToken = obj.Token
	}
	if len(obj.DiscoveryToken) == 0 && len(obj.DiscoveryFile) == 0 {
		obj.DiscoveryToken = obj.Token
	}
	// Make sure file URLs become paths
	if len(obj.DiscoveryFile) != 0 {
		u, err := url.Parse(obj.DiscoveryFile)
		if err == nil && u.Scheme == "file" {
			obj.DiscoveryFile = u.Path
		}
	}
}

// SetDefaultsEtcdSelfHosted sets defaults for self-hosted etcd
func SetDefaultsEtcdSelfHosted(obj *MasterConfiguration) {
	if obj.Etcd.SelfHosted == nil {
		obj.Etcd.SelfHosted = &SelfHostedEtcd{}
	}

	if obj.Etcd.SelfHosted.ClusterServiceName == "" {
		obj.Etcd.SelfHosted.ClusterServiceName = DefaultEtcdClusterServiceName
	}

	if obj.Etcd.SelfHosted.EtcdVersion == "" {
		obj.Etcd.SelfHosted.EtcdVersion = constants.DefaultEtcdVersion
	}

	if obj.Etcd.SelfHosted.OperatorVersion == "" {
		obj.Etcd.SelfHosted.OperatorVersion = DefaultEtcdOperatorVersion
	}

	if obj.Etcd.SelfHosted.CertificatesDir == "" {
		obj.Etcd.SelfHosted.CertificatesDir = DefaultEtcdCertDir
	}
}

// SetDefaults_KubeletConfiguration assigns default values to kubelet
func SetDefaults_KubeletConfiguration(obj *MasterConfiguration) {
	if obj.KubeletConfiguration.BaseConfig == nil {
		obj.KubeletConfiguration.BaseConfig = &kubeletconfigv1alpha1.KubeletConfiguration{}
	}
	if obj.KubeletConfiguration.BaseConfig.PodManifestPath == "" {
		obj.KubeletConfiguration.BaseConfig.PodManifestPath = DefaultManifestsDir
	}
	if obj.KubeletConfiguration.BaseConfig.AllowPrivileged == nil {
		obj.KubeletConfiguration.BaseConfig.AllowPrivileged = utilpointer.BoolPtr(true)
	}
	if obj.KubeletConfiguration.BaseConfig.ClusterDNS == nil {
		dnsIP, err := constants.GetDNSIP(obj.Networking.ServiceSubnet)
		if err != nil {
			obj.KubeletConfiguration.BaseConfig.ClusterDNS = []string{DefaultClusterDNSIP}
		} else {
			obj.KubeletConfiguration.BaseConfig.ClusterDNS = []string{dnsIP.String()}
		}
	}
	if obj.KubeletConfiguration.BaseConfig.ClusterDomain == "" {
		obj.KubeletConfiguration.BaseConfig.ClusterDomain = DefaultServiceDNSDomain
	}
	if obj.KubeletConfiguration.BaseConfig.Authorization.Mode == "" {
		obj.KubeletConfiguration.BaseConfig.Authorization.Mode = kubeletconfigv1alpha1.KubeletAuthorizationModeWebhook
	}
	if obj.KubeletConfiguration.BaseConfig.Authentication.X509.ClientCAFile == "" {
		obj.KubeletConfiguration.BaseConfig.Authentication.X509.ClientCAFile = DefaultCACertPath
	}
	if obj.KubeletConfiguration.BaseConfig.CAdvisorPort == nil {
		obj.KubeletConfiguration.BaseConfig.CAdvisorPort = utilpointer.Int32Ptr(0)
	}
}
