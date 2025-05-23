/*
Copyright 2022 Red Hat
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

package helpers

import (
	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetSecret fetches a Secret resource
//
// Example usage:
//
//	secret := th.GetSecret(types.NamespacedName{Name: "test-secret", Namespace: "test-namespace"})
func (tc *TestHelper) GetSecret(name types.NamespacedName) corev1.Secret {
	secret := &corev1.Secret{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, secret)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return *secret
}

// GetSecrets retrieves all secrets by namespace
func (tc *TestHelper) GetSecrets(namespace string) *corev1.SecretList {
	scerets := &corev1.SecretList{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.List(tc.Ctx, scerets, client.InNamespace(namespace))).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return scerets
}

// CreateSecret creates a new Secret resource with provided data.
//
// Example usage:
//
//	secret := th.CreateSecret(types.NamespacedName{Name: "test-secret", Namespace: "test-namespace"}, map[string][]byte{"key": []byte("value")})
func (tc *TestHelper) CreateSecret(name types.NamespacedName, data map[string][]byte) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Data: data,
	}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Create(tc.Ctx, secret)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return secret
}

// CreateEmptySecret creates a new empty Secret resource .
//
// Example usage:
//
//	secret := th.CreateSecret(types.NamespacedName{Name: "test-secret", Namespace: "test-namespace"})
func (tc *TestHelper) CreateEmptySecret(name types.NamespacedName) *corev1.Secret {
	return tc.CreateSecret(name, map[string][]byte{})
}

// CreateSSHSecret creates new secret containing mock "ssh-privatekey"
//
// Example usage:
//
// sshSecret := th.CreateSSHSecret(types.NamespacedName{Name: "test-ssh-secret", Namespace: "test-namespace"})
func (tc *TestHelper) CreateSSHSecret(name types.NamespacedName) *corev1.Secret {
	return tc.CreateSecret(name, map[string][]byte{"ssh-privatekey": []byte("foo")})
}

// DeleteSecret deletes a Secret resource
//
// Example usage:
//
//	CreateNovaExternalComputeSSHSecret(sshSecretName)
//	DeferCleanup(th.DeleteSecret, sshSecretName)
func (tc *TestHelper) DeleteSecret(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		secret := &corev1.Secret{}
		err := tc.K8sClient.Get(tc.Ctx, name, secret)
		// if it is already gone that is OK
		if k8s_errors.IsNotFound(err) {
			return
		}
		g.Expect(err).ShouldNot(gomega.HaveOccurred())

		g.Expect(tc.K8sClient.Delete(tc.Ctx, secret)).Should(gomega.Succeed())

		err = tc.K8sClient.Get(tc.Ctx, name, secret)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}

// AssertSecretDoesNotExist ensures the Secret resource does not exist in a k8s cluster.
func (tc *TestHelper) AssertSecretDoesNotExist(name types.NamespacedName) {
	instance := &corev1.Secret{}
	gomega.Eventually(func(g gomega.Gomega) {
		err := tc.K8sClient.Get(tc.Ctx, name, instance)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}

// UpdateSecret adds a new key or updates an existing key in the given Secret with a new value
func (tc *TestHelper) UpdateSecret(secretName types.NamespacedName, key string, newValue []byte) {
	gomega.Eventually(func(g gomega.Gomega) {
		secret := tc.GetSecret(secretName)
		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		}
		secret.Data[key] = newValue
		g.Expect(tc.K8sClient.Update(tc.Ctx, &secret)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
	tc.Logger.Info("Secret updated", "secret", secretName, "key", key)
}
