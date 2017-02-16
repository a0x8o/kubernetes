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

package util

import (
	"bytes"
	"testing"
	"time"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

func TestTokenParse(t *testing.T) {
	var tests = []struct {
		token    string
		expected bool
	}{
		{token: "1234567890123456789012", expected: false},   // invalid parcel size
		{token: "12345:1234567890123456", expected: false},   // invalid parcel size
		{token: ".1234567890123456", expected: false},        // invalid parcel size
		{token: "123456:1234567890.123456", expected: false}, // invalid separation
		{token: "abcdef.1234567890123456", expected: false},  // invalid separation
		{token: "Abcdef:1234567890123456", expected: false},  // invalid token id
		{token: "123456:AABBCCDDEEFFGGHH", expected: false},  // invalid token secret
		{token: "abcdef:1234567890123456", expected: true},
		{token: "123456:aabbccddeeffgghh", expected: true},
	}

	for _, rt := range tests {
		_, _, actual := ParseToken(rt.token)
		if (actual == nil) != rt.expected {
			t.Errorf(
				"failed ParseToken for this token: [%s]\n\texpected: %t\n\t  actual: %t",
				rt.token,
				rt.expected,
				(actual == nil),
			)
		}
	}

}

func TestParseTokenID(t *testing.T) {
	var tests = []struct {
		tokenID  string
		expected bool
	}{
		{tokenID: "", expected: false},
		{tokenID: "1234567890123456789012", expected: false},
		{tokenID: "12345", expected: false},
		{tokenID: "Abcdef", expected: false},
		{tokenID: "abcdef", expected: true},
		{tokenID: "123456", expected: true},
	}
	for _, rt := range tests {
		actual := ParseTokenID(rt.tokenID)
		if (actual == nil) != rt.expected {
			t.Errorf(
				"failed ParseTokenID for this token ID: [%s]\n\texpected: %t\n\t  actual: %t",
				rt.tokenID,
				rt.expected,
				(actual == nil),
			)
		}
	}
}

func TestValidateToken(t *testing.T) {
	var tests = []struct {
		token    *kubeadmapi.TokenDiscovery
		expected bool
	}{
		{token: &kubeadmapi.TokenDiscovery{ID: "", Secret: ""}, expected: false},
		{token: &kubeadmapi.TokenDiscovery{ID: "1234567890123456789012", Secret: ""}, expected: false},
		{token: &kubeadmapi.TokenDiscovery{ID: "", Secret: "1234567890123456789012"}, expected: false},
		{token: &kubeadmapi.TokenDiscovery{ID: "12345", Secret: "1234567890123456"}, expected: false},
		{token: &kubeadmapi.TokenDiscovery{ID: "Abcdef", Secret: "1234567890123456"}, expected: false},
		{token: &kubeadmapi.TokenDiscovery{ID: "123456", Secret: "AABBCCDDEEFFGGHH"}, expected: false},
		{token: &kubeadmapi.TokenDiscovery{ID: "abc*ef", Secret: "1234567890123456"}, expected: false},
		{token: &kubeadmapi.TokenDiscovery{ID: "abcdef", Secret: "123456789*123456"}, expected: false},
		{token: &kubeadmapi.TokenDiscovery{ID: "abcdef", Secret: "1234567890123456"}, expected: true},
		{token: &kubeadmapi.TokenDiscovery{ID: "123456", Secret: "aabbccddeeffgghh"}, expected: true},
		{token: &kubeadmapi.TokenDiscovery{ID: "abc456", Secret: "1234567890123456"}, expected: true},
		{token: &kubeadmapi.TokenDiscovery{ID: "abcdef", Secret: "123456ddeeffgghh"}, expected: true},
	}
	for _, rt := range tests {
		valid, actual := ValidateToken(rt.token)
		if (actual == nil) != rt.expected {
			t.Errorf(
				"failed ValidateToken for this token ID: [%s]\n\texpected: %t\n\t  actual: %t",
				rt.token,
				rt.expected,
				(actual == nil),
			)
		}
		if (valid == true) != rt.expected {
			t.Errorf(
				"failed ValidateToken for this token ID: [%s]\n\texpected: %t\n\t  actual: %t",
				rt.token,
				rt.expected,
				(actual == nil),
			)
		}
	}
}

func TestGenerateToken(t *testing.T) {
	td := &kubeadmapi.TokenDiscovery{}
	if err := GenerateToken(td); err != nil {
		t.Fatalf("GenerateToken returned an unexpected error: %+v", err)
	}
	if len(td.ID) != 6 {
		t.Errorf("failed GenerateToken first part length:\n\texpected: 6\n\t  actual: %d", len(td.ID))
	}
	if len(td.Secret) != 16 {
		t.Errorf("failed GenerateToken second part length:\n\texpected: 16\n\t  actual: %d", len(td.Secret))
	}
}

func TestRandBytes(t *testing.T) {
	var randTest = []int{
		0,
		1,
		2,
		3,
		100,
	}

	for _, rt := range randTest {
		actual, err := randBytes(rt)
		if err != nil {
			t.Errorf("failed randBytes: %v", err)
		}
		if len(actual) != rt*2 {
			t.Errorf("failed randBytes:\n\texpected: %d\n\t  actual: %d\n", rt*2, len(actual))
		}
	}
}

func TestDiscoveryPort(t *testing.T) {
	var tests = []struct {
		token    *kubeadmapi.TokenDiscovery
		expected int32
	}{
		{token: &kubeadmapi.TokenDiscovery{}, expected: 9898}, // should use default
		{token: &kubeadmapi.TokenDiscovery{Addresses: []string{"foobar:1234"}}, expected: 1234},
		{token: &kubeadmapi.TokenDiscovery{Addresses: []string{"doesnothaveport"}}, expected: 9898}, // should use default
		{token: &kubeadmapi.TokenDiscovery{Addresses: []string{"foorbar:abcd"}}, expected: 9898},    // since abcd isn't an int, should use default
	}
	for _, rt := range tests {
		actual := DiscoveryPort(rt.token)
		if actual != rt.expected {
			t.Errorf(
				"failed DiscoveryPort:\n\texpected: %d\n\t  actual: %d",
				rt.expected,
				actual,
			)
		}
	}
}

func TestBearerToken(t *testing.T) {
	var tests = []struct {
		token    *kubeadmapi.TokenDiscovery
		expected string
	}{
		{token: &kubeadmapi.TokenDiscovery{ID: "foo", Secret: "bar"}, expected: "foo:bar"}, // should use default
	}
	for _, rt := range tests {
		actual := BearerToken(rt.token)
		if actual != rt.expected {
			t.Errorf(
				"failed BearerToken:\n\texpected: %s\n\t  actual: %s",
				rt.expected,
				actual,
			)
		}
	}
}

func TestEncodeTokenSecretData(t *testing.T) {
	var tests = []struct {
		token *kubeadmapi.TokenDiscovery
		t     time.Duration
	}{
		{token: &kubeadmapi.TokenDiscovery{ID: "foo", Secret: "bar"}},                 // should use default
		{token: &kubeadmapi.TokenDiscovery{ID: "foo", Secret: "bar"}, t: time.Second}, // should use default
	}
	for _, rt := range tests {
		actual := encodeTokenSecretData(rt.token, rt.t)
		if !bytes.Equal(actual["token-id"], []byte(rt.token.ID)) {
			t.Errorf(
				"failed EncodeTokenSecretData:\n\texpected: %s\n\t  actual: %s",
				rt.token.ID,
				actual["token-id"],
			)
		}
		if !bytes.Equal(actual["token-secret"], []byte(rt.token.Secret)) {
			t.Errorf(
				"failed EncodeTokenSecretData:\n\texpected: %s\n\t  actual: %s",
				rt.token.Secret,
				actual["token-secret"],
			)
		}
		if rt.t > 0 {
			if actual["expiration"] == nil {
				t.Errorf(
					"failed EncodeTokenSecretData, duration was not added to time",
				)
			}
		}
	}
}
