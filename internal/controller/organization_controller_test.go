package controller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
)

var _ = Describe("Organization Controller", func() {
	const (
		organizationName      = "test-organization"
		organizationNamespace = "default"
		timeout               = 120 * time.Second
		interval              = 2 * time.Second
	)

	Context("When reconciling an Organization", func() {
		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      organizationName,
			Namespace: organizationNamespace,
		}

		BeforeEach(func() {
			organization := &securityv1alpha1.Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name:      organizationName,
					Namespace: organizationNamespace,
				},
				Spec: securityv1alpha1.OrganizationSpec{
					Name: organizationName,
				},
			}
			Expect(k8sClient.Create(ctx, organization)).To(Succeed())
			fmt.Printf("Organization created: %+v\n", organization)
		})

		AfterEach(func() {
			organization := &securityv1alpha1.Organization{}
			err := k8sClient.Get(ctx, typeNamespacedName, organization)
			if err == nil {
				fmt.Printf("Cleaning up organization: %+v\n", organization)
				Expect(k8sClient.Delete(ctx, organization)).To(Succeed())
			} else {
				fmt.Printf("Error getting organization during cleanup: %v\n", err)
			}
		})

		It("should delete the namespace when the organization is deleted", func() {
			organizationReconciler := &OrganizationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			By("Triggering initial reconciliation")
			result, err := organizationReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("Initial reconciliation result: %+v\n", result)

			By("Verifying the namespace was created")
			createdNamespace := &corev1.Namespace{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("org-%s", organizationName)}, createdNamespace)
			}, timeout, interval).Should(Succeed())
			fmt.Printf("Namespace created: %+v\n", createdNamespace)

			By("Deleting the organization")
			organization := &securityv1alpha1.Organization{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, organization)).To(Succeed())
			fmt.Printf("Organization before deletion: %+v\n", organization)
			Expect(k8sClient.Delete(ctx, organization)).To(Succeed())
			fmt.Println("Delete request sent for organization")

			By("Triggering reconciliation after deletion")
			result, err = organizationReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("Reconciliation result after deletion: %+v\n", result)

			By("Checking if the namespace was deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: "org-" + organizationName}, &corev1.Namespace{})
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue(), "Namespace should be deleted")
		})
	})
})
