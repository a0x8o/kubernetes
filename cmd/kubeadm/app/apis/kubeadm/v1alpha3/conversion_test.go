/*
Copyright 2018 The Kubernetes Authors.

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

package v1alpha3_test

import (
	"testing"

	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1alpha3"
	testutil "k8s.io/kubernetes/cmd/kubeadm/test"
)

func TestJoinConfigurationConversion(t *testing.T) {
	testcases := map[string]struct {
		old         *v1alpha3.JoinConfiguration
		expectedErr string
	}{
		"conversion succeeds": {
			old:         &v1alpha3.JoinConfiguration{},
			expectedErr: "",
		},
		"cluster name fails to be converted": {
			old: &v1alpha3.JoinConfiguration{
				ClusterName: "kubernetes",
			},
			expectedErr: "clusterName has been removed from JoinConfiguration and clusterName from ClusterConfiguration will be used instead. Please cleanup JoinConfiguration.ClusterName fields",
		},
		"feature gates fails to be converted": {
			old: &v1alpha3.JoinConfiguration{
				FeatureGates: map[string]bool{
					"someGate": true,
				},
			},
			expectedErr: "featureGates has been removed from JoinConfiguration and featureGates from ClusterConfiguration will be used instead. Please cleanup JoinConfiguration.FeatureGates fields",
		},
	}
	for _, tc := range testcases {
		internal := &kubeadm.JoinConfiguration{}
		err := scheme.Scheme.Convert(tc.old, internal, nil)
		if len(tc.expectedErr) != 0 {
			testutil.AssertError(t, err, tc.expectedErr)
		} else if err != nil {
			t.Errorf("no error was expected but '%s' was found", err)
		}
	}
}

func TestConvertToUseHyperKubeImage(t *testing.T) {
	tests := []struct {
		desc              string
		in                *v1alpha3.ClusterConfiguration
		useHyperKubeImage bool
		expectedErr       bool
	}{
		{
			desc:              "unset UnifiedControlPlaneImage sets UseHyperKubeImage to false",
			in:                &v1alpha3.ClusterConfiguration{},
			useHyperKubeImage: false,
			expectedErr:       false,
		},
		{
			desc: "matching UnifiedControlPlaneImage sets UseHyperKubeImage to true",
			in: &v1alpha3.ClusterConfiguration{
				ImageRepository:          "k8s.gcr.io",
				KubernetesVersion:        "v1.12.2",
				UnifiedControlPlaneImage: "k8s.gcr.io/hyperkube:v1.12.2",
			},
			useHyperKubeImage: true,
			expectedErr:       false,
		},
		{
			desc: "mismatching UnifiedControlPlaneImage tag causes an error",
			in: &v1alpha3.ClusterConfiguration{
				ImageRepository:          "k8s.gcr.io",
				KubernetesVersion:        "v1.12.0",
				UnifiedControlPlaneImage: "k8s.gcr.io/hyperkube:v1.12.2",
			},
			expectedErr: true,
		},
		{
			desc: "mismatching UnifiedControlPlaneImage repo causes an error",
			in: &v1alpha3.ClusterConfiguration{
				ImageRepository:          "my.repo",
				KubernetesVersion:        "v1.12.2",
				UnifiedControlPlaneImage: "k8s.gcr.io/hyperkube:v1.12.2",
			},
			expectedErr: true,
		},
		{
			desc: "mismatching UnifiedControlPlaneImage image name causes an error",
			in: &v1alpha3.ClusterConfiguration{
				ImageRepository:          "k8s.gcr.io",
				KubernetesVersion:        "v1.12.2",
				UnifiedControlPlaneImage: "k8s.gcr.io/otherimage:v1.12.2",
			},
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			out := &kubeadm.ClusterConfiguration{}
			err := v1alpha3.Convert_v1alpha3_UnifiedControlPlaneImage_To_kubeadm_UseHyperKubeImage(test.in, out)
			if test.expectedErr {
				if err == nil {
					t.Fatalf("unexpected success, UseHyperKubeImage: %t", out.UseHyperKubeImage)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected failure: %v", err)
				}
				if out.UseHyperKubeImage != test.useHyperKubeImage {
					t.Fatalf("mismatching result from conversion:\n\tExpected: %t\n\tReceived: %t", test.useHyperKubeImage, out.UseHyperKubeImage)
				}
			}
		})
	}
}
