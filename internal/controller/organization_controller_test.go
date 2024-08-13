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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
)

var _ = Describe("Organization controller", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating and deleting Organizations", func() {
		It("Should remove the finalizer when deleting an Organization", func() {
			ctx := context.Background()

			org := &securityv1alpha1.Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-finalizer",
					Finalizers: []string{"organization.giantswarm.io/finalizer"},
				},
				Spec: securityv1alpha1.OrganizationSpec{},
				Status: securityv1alpha1.OrganizationStatus{
					Namespace: "org-test-finalizer",
				},
			}
			Expect(k8sClient.Create(ctx, org)).Should(Succeed())

			// Create the associated namespace
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "org-test-finalizer",
				},
			}
			Expect(k8sClient.Create(ctx, namespace)).Should(Succeed())

			reconciler := &OrganizationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Trigger deletion
			Expect(k8sClient.Delete(ctx, org)).Should(Succeed())

			// Wait for the organization to be fully deleted
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "test-finalizer"}, &securityv1alpha1.Organization{})
				if errors.IsNotFound(err) {
					return nil
				}
				if err != nil {
					return err
				}
				// Trigger reconciliation if the organization still exists
				_, reconcileErr := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{Name: "test-finalizer"},
				})
				if reconcileErr != nil {
					return reconcileErr
				}
				return fmt.Errorf("organization still exists")
			}, timeout, interval).Should(Succeed())

			// Verify that the namespace has been deleted
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "org-test-finalizer"}, &corev1.Namespace{})
				if errors.IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("namespace still exists")
			}, timeout, interval).Should(Succeed())

			// Verify that the organization count metric has been updated
			Expect(testutil.ToFloat64(organizationsTotal)).Should(Equal(float64(0)))
		})

	})
})
