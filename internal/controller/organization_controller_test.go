/*
Copyright 2024.

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

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
)

var _ = Describe("Organization controller", func() {
	const (
		OrganizationName = "test-org"
		timeout          = time.Second * 10
		interval         = time.Millisecond * 250
	)

	Context("When creating an Organization", func() {
		It("Should create a corresponding Namespace and update the Organization status", func() {
			ctx := context.Background()
			organization := &securityv1alpha1.Organization{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "security.giantswarm.io/v1alpha1",
					Kind:       "Organization",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: OrganizationName,
				},
				Spec: securityv1alpha1.OrganizationSpec{},
			}

			// Create the Organization
			Expect(k8sClient.Create(ctx, organization)).Should(Succeed())

			// Set up the OrganizationReconciler
			reconciler := &OrganizationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Reconcile
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: OrganizationName},
			})
			Expect(err).NotTo(HaveOccurred())

			// Check if the Namespace was created
			namespaceName := "org-" + OrganizationName
			createdNamespace := &corev1.Namespace{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: namespaceName}, createdNamespace)
			}, timeout, interval).Should(Succeed())

			Expect(createdNamespace.Labels).To(HaveKeyWithValue("giantswarm.io/organization", OrganizationName))
			Expect(createdNamespace.Labels).To(HaveKeyWithValue("giantswarm.io/managed-by", "organization-operator"))

			// Verify the Organization status was updated
			updatedOrg := &securityv1alpha1.Organization{}
			Eventually(func() string {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: OrganizationName}, updatedOrg)
				if err != nil {
					return ""
				}
				return updatedOrg.Status.Namespace
			}, timeout, interval).Should(Equal(namespaceName))
		})
	})
})
