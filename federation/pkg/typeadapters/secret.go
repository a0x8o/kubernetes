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

package typeadapters

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	federationclientset "k8s.io/kubernetes/federation/client/clientset_generated/federation_clientset"
	"k8s.io/kubernetes/federation/pkg/federation-controller/util"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

type SecretAdapter struct {
	client federationclientset.Interface
}

func NewSecretAdapter(client federationclientset.Interface) *SecretAdapter {
	return &SecretAdapter{client: client}
}

func (a *SecretAdapter) SetClient(client federationclientset.Interface) {
	a.client = client
}

func (a *SecretAdapter) Kind() string {
	return "secret"
}

func (a *SecretAdapter) Equivalent(obj1, obj2 pkgruntime.Object) bool {
	secret1 := obj1.(*apiv1.Secret)
	secret2 := obj2.(*apiv1.Secret)
	return util.SecretEquivalent(*secret1, *secret2)
}

func (a *SecretAdapter) ObjectMeta(obj pkgruntime.Object) *metav1.ObjectMeta {
	return &obj.(*apiv1.Secret).ObjectMeta
}

func (a *SecretAdapter) NamespacedName(obj pkgruntime.Object) types.NamespacedName {
	secret := obj.(*apiv1.Secret)
	return types.NamespacedName{Namespace: secret.Namespace, Name: secret.Name}
}

func (a *SecretAdapter) FedCreate(obj pkgruntime.Object) (pkgruntime.Object, error) {
	secret := obj.(*apiv1.Secret)
	return a.client.CoreV1().Secrets(secret.Namespace).Create(secret)
}

func (a *SecretAdapter) FedGet(namespacedName types.NamespacedName) (pkgruntime.Object, error) {
	return a.client.CoreV1().Secrets(namespacedName.Namespace).Get(namespacedName.Name, metav1.GetOptions{})
}

func (a *SecretAdapter) FedUpdate(obj pkgruntime.Object) (pkgruntime.Object, error) {
	secret := obj.(*apiv1.Secret)
	return a.client.CoreV1().Secrets(secret.Namespace).Update(secret)
}

func (a *SecretAdapter) FedDelete(namespacedName types.NamespacedName, options *metav1.DeleteOptions) error {
	return a.client.CoreV1().Secrets(namespacedName.Namespace).Delete(namespacedName.Name, options)
}

func (a *SecretAdapter) ClusterGet(client clientset.Interface, namespacedName types.NamespacedName) (pkgruntime.Object, error) {
	return client.CoreV1().Secrets(namespacedName.Namespace).Get(namespacedName.Name, metav1.GetOptions{})
}

func (a *SecretAdapter) NewTestObject(namespace string) pkgruntime.Object {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-secret-",
			Namespace:    namespace,
		},
		Data: map[string][]byte{
			"A": []byte("ala ma kota"),
		},
		Type: apiv1.SecretTypeOpaque,
	}
}
