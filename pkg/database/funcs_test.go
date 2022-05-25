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

package database

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"
)

//
// CreateOrPatchDB - create or patch the service DB instance
//
func TestCreateOrPatchDB(t *testing.T) {
	t.Run("Create database", func(t *testing.T) {
		g := NewWithT(t)
		clientBuilder := fake.NewClientBuilder()

		dbOptions := Options{
			DatabaseHostname: "dbhost",
			DatabaseName:     "dbname",
			Secret:           "dbsecret",
		}

		scheme := runtime.NewScheme()
		_ = keystonev1.AddToScheme(scheme)

		obj := &keystonev1.KeystoneAPI{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "keystone",
				Namespace: "openstack",
			},
		}

		output := &mariadbv1.MariaDBDatabase{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "keystone",
				Namespace: "openstack",
			},
		}

		db, _, _, _ := CreateOrPatchDB(context.TODO(), clientBuilder.Build(), obj, scheme, dbOptions)

		g.Expect(db).To(Equal(output))

		// TODO improve tests using ginkgo/envtest to actually create the object and fetch it that we can check the spec
		// something like that, not tested - https://book.kubebuilder.io/reference/envtest.html
		/*
			By("returning no error")

			Expect(err).NotTo(HaveOccurred())

			By("returning OperationResultCreated")
			Expect(op).To(BeEquivalentTo(controllerutil.OperationResultCreated))
			By("actually having the DB created")
			fetched := &mariadbv1.MariaDBDatabase{}
			Expect(
				c.Get(context.TODO(),
					types.NamespacedName{
						Name:      "keystone-keystone",
						Namespace: "default",
					},
					fetched)).To(Succeed())
			Expect(db).To(Equal(output))

			By("spec being mutated by MutateFn")
			Expect(fetched.Spec.Name).To(Equal("dbname"))
			Expect(fetched.Spec.Secret).To(Equal("dbsecret"))
		*/
	})
}
