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

package api

import (
	"net"
)

// KubeadmConfig TODO add description
// TODO(phase1+) @krousey: Please don't embed structs. It obfuscates the source of the fields and doesn't really buy you anything.
type KubeadmConfig struct {
	InitFlags
	JoinFlags
	Secrets struct {
		GivenToken  string // dot-separated `<TokenID>.<Token>` set by the user
		TokenID     string // optional on master side, will be generated if not specified
		Token       []byte // optional on master side, will be generated if not specified
		BearerToken string // set based on Token
	}
	EnvParams map[string]string // TODO(phase2) this is likely to be come componentconfig
}

// TODO(phase2) should we add validation functions for these structs?

// TODO(phase1+) refactor token handling
// - https://github.com/kubernetes/kubernetes/pull/33262/files#r80333662
// - https://github.com/kubernetes/kubernetes/pull/33262/files#r80336374
// - https://github.com/kubernetes/kubernetes/pull/33262/files#r80333982

// InitFlags holds values for "kubeadm init" command flags.
type InitFlags struct {
	API struct {
		AdvertiseAddrs   []net.IP
		ExternalDNSNames []string
		Etcd             struct {
			ExternalEndpoints []string
			ExternalCAFile    string
			ExternalCertFile  string
			ExternalKeyFile   string
		}
	}
	Services struct {
		CIDR      net.IPNet
		DNSDomain string
	}
	PodNetwork struct {
		CIDR net.IPNet
	}
	Versions struct {
		Kubernetes string
	}
	CloudProvider string
}

const (
	DefaultServiceDNSDomain   = "cluster.local"
	DefaultServicesCIDRString = "10.12.0.0/12"
	DefaultKubernetesVersion  = "v1.4.0"
)

var (
	DefaultServicesCIDR  *net.IPNet
	ListOfCloudProviders = []string{
		"aws",
		"azure",
		"cloudstack",
		"gce",
		"mesos",
		"openstack",
		"ovirt",
		"rackspace",
		"vsphere",
	}
)

func init() {
	_, DefaultServicesCIDR, _ = net.ParseCIDR(DefaultServicesCIDRString)
}

// JoinFlags holds values for "kubeadm join" command flags.
type JoinFlags struct {
	MasterAddrs []net.IP
	// TODO(phase1+) add manual mode flags here, e.g. RootCACertPath
}

// ClusterInfo TODO add description
type ClusterInfo struct {
	// TODO(phase1+) this may become simply `api.Config`
	CertificateAuthorities []string `json:"certificateAuthorities"`
	Endpoints              []string `json:"endpoints"`
}
