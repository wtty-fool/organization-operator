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
		timeout               = 5 * time.Minute
		interval              = 250 * time.Millisecond
	)

	Context("When reconciling an Organization", func() {
		var (
			ctx                context.Context
			typeNamespacedName types.NamespacedName
			reconciler         *OrganizationReconciler
		)

		BeforeEach(func() {
			ctx = context.Background()
			typeNamespacedName = types.NamespacedName{
				Name:      organizationName,
				Namespace: organizationNamespace,
			}
			reconciler = &OrganizationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Clean up any existing organization before each test
			cleanupExistingOrganization(ctx, typeNamespacedName)

			// Create a new organization
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
		})

		AfterEach(func() {
			cleanupExistingOrganization(ctx, typeNamespacedName)
		})

		It("should create a namespace and update the organization status", func() {
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the namespace was created")
			createdNamespace := &corev1.Namespace{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("org-%s", organizationName)}, createdNamespace)
			}, timeout, interval).Should(Succeed())

			By("Verifying the organization status was updated")
			updatedOrg := &securityv1alpha1.Organization{}
			Eventually(func() string {
				err := k8sClient.Get(ctx, typeNamespacedName, updatedOrg)
				if err != nil {
					return ""
				}
				return updatedOrg.Status.Namespace
			}, timeout, interval).Should(Equal(fmt.Sprintf("org-%s", organizationName)))
		})

		It("should add a finalizer to the organization", func() {
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())

			updatedOrg := &securityv1alpha1.Organization{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, updatedOrg)
				if err != nil {
					return false
				}
				return hasOrganizationFinalizer(updatedOrg)
			}, timeout, interval).Should(BeTrue())
		})

		//It("should delete the namespace when the organization is deleted", func() {
		//	_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
		//	Expect(err).NotTo(HaveOccurred())
		//
		//	organization := &securityv1alpha1.Organization{}
		//	Expect(k8sClient.Get(ctx, typeNamespacedName, organization)).To(Succeed())
		//	Expect(k8sClient.Delete(ctx, organization)).To(Succeed())
		//
		//	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
		//	Expect(err).NotTo(HaveOccurred())
		//
		//	By("Checking if the namespace was deleted")
		//	Eventually(func() bool {
		//		err := k8sClient.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("org-%s", organizationName)}, &corev1.Namespace{})
		//		return errors.IsNotFound(err)
		//	}, timeout, interval).Should(BeTrue(), "Namespace should be deleted")
		//
		//	By("Checking if the organization was deleted")
		//	Eventually(func() bool {
		//		err := k8sClient.Get(ctx, typeNamespacedName, &securityv1alpha1.Organization{})
		//		return errors.IsNotFound(err)
		//	}, timeout, interval).Should(BeTrue(), "Organization should be deleted")
		//})
	})
})

func cleanupExistingOrganization(ctx context.Context, namespacedName types.NamespacedName) {
	existingOrg := &securityv1alpha1.Organization{}
	err := k8sClient.Get(ctx, namespacedName, existingOrg)
	if err == nil {
		if hasOrganizationFinalizer(existingOrg) {
			existingOrg.ObjectMeta.Finalizers = removeOrganizationFinalizer(existingOrg.ObjectMeta.Finalizers)
			err = k8sClient.Update(ctx, existingOrg)
			Expect(err).NotTo(HaveOccurred())
		}

		err = k8sClient.Delete(ctx, existingOrg)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, namespacedName, existingOrg)
			return errors.IsNotFound(err)
		}, 1*time.Minute, 250*time.Millisecond).Should(BeTrue(), "Organization should be deleted during cleanup")
	}
}
