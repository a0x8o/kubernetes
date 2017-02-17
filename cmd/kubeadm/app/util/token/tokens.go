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

package token

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

const (
	TokenIDBytes     = 3
	TokenSecretBytes = 8
)

var (
	tokenIDRegexpString = "^([a-z0-9]{6})$"
	tokenIDRegexp       = regexp.MustCompile(tokenIDRegexpString)
	tokenRegexpString   = "^([a-z0-9]{6})\\:([a-z0-9]{16})$"
	tokenRegexp         = regexp.MustCompile(tokenRegexpString)
)

func randBytes(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateToken generates a new token with a token ID that is valid as a
// Kubernetes DNS label.
// For more info, see kubernetes/pkg/util/validation/validation.go.
func GenerateToken(d *kubeadmapi.TokenDiscovery) error {
	tokenID, err := randBytes(TokenIDBytes)
	if err != nil {
		return err
	}

	token, err := randBytes(TokenSecretBytes)
	if err != nil {
		return err
	}

	d.ID = strings.ToLower(tokenID)
	d.Secret = strings.ToLower(token)
	return nil
}

// ParseTokenID tries and parse a valid token ID from a string.
// An error is returned in case of failure.
func ParseTokenID(s string) error {
	if !tokenIDRegexp.MatchString(s) {
		return fmt.Errorf("token ID [%q] was not of form [%q]", s, tokenIDRegexpString)
	}
	return nil
}

// ParseToken tries and parse a valid token from a string.
// A token ID and token secret are returned in case of success, an error otherwise.
func ParseToken(s string) (string, string, error) {
	split := tokenRegexp.FindStringSubmatch(s)
	if len(split) != 3 {
		return "", "", fmt.Errorf("token [%q] was not of form [%q]", s, tokenRegexpString)
	}
	return split[1], split[2], nil
}

// BearerToken returns a string representation of the passed token.
func BearerToken(d *kubeadmapi.TokenDiscovery) string {
	return fmt.Sprintf("%s:%s", d.ID, d.Secret)
}

// ValidateToken validates whether a token is well-formed.
// In case it's not, the corresponding error is returned as well.
func ValidateToken(d *kubeadmapi.TokenDiscovery) (bool, error) {
	if _, _, err := ParseToken(d.ID + ":" + d.Secret); err != nil {
		return false, err
	}
	return true, nil
}
