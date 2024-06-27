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
	"sigs.k8s.io/controller-runtime/pkg/log"
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
		ctx := context.Background()
		logger := log.FromContext(ctx)

		typeNamespacedName := types.NamespacedName{
			Name:      organizationName,
			Namespace: organizationNamespace,
		}

		BeforeEach(func() {
			// Clean up any existing organization before each test
			existingOrg := &securityv1alpha1.Organization{}
			err := k8sClient.Get(ctx, typeNamespacedName, existingOrg)
			if err == nil {
				logger.Info("Deleting existing organization", "name", existingOrg.Name)

				// Remove finalizer if it exists
				if containsString(existingOrg.ObjectMeta.Finalizers, organizationFinalizerName) {
					existingOrg.ObjectMeta.Finalizers = removeString(existingOrg.ObjectMeta.Finalizers, organizationFinalizerName)
					err = k8sClient.Update(ctx, existingOrg)
					Expect(err).NotTo(HaveOccurred())
				}

				err = k8sClient.Delete(ctx, existingOrg)
				Expect(err).NotTo(HaveOccurred())
			} else if !errors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			// Wait for the organization to be deleted
			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, existingOrg)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

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
			// Clean up the organization after each test
			organization := &securityv1alpha1.Organization{}
			err := k8sClient.Get(ctx, typeNamespacedName, organization)
			if err == nil {
				// Remove finalizer if it exists
				if containsString(organization.ObjectMeta.Finalizers, organizationFinalizerName) {
					organization.ObjectMeta.Finalizers = removeString(organization.ObjectMeta.Finalizers, organizationFinalizerName)
					err = k8sClient.Update(ctx, organization)
					Expect(err).NotTo(HaveOccurred())
				}

				err = k8sClient.Delete(ctx, organization)
				Expect(err).NotTo(HaveOccurred())

				// Wait for the organization to be deleted
				Eventually(func() bool {
					err := k8sClient.Get(ctx, typeNamespacedName, organization)
					return errors.IsNotFound(err)
				}, timeout, interval).Should(BeTrue())
			}
		})

		It("should create a namespace and update the organization status", func() {
			organizationReconciler := &OrganizationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := organizationReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())

			createdNamespace := &corev1.Namespace{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("org-%s", organizationName)}, createdNamespace)
			}, timeout, interval).Should(Succeed())

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
			organizationReconciler := &OrganizationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := organizationReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())

			updatedOrg := &securityv1alpha1.Organization{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, updatedOrg)
				if err != nil {
					return false
				}
				return containsString(updatedOrg.ObjectMeta.Finalizers, organizationFinalizerName)
			}, timeout, interval).Should(BeTrue())
		})

		It("should delete the namespace when the organization is deleted", func() {
			organizationReconciler := &OrganizationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := organizationReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())

			organization := &securityv1alpha1.Organization{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, organization)).To(Succeed())
			Expect(k8sClient.Delete(ctx, organization)).To(Succeed())

			_, err = organizationReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("org-%s", organizationName)}, &corev1.Namespace{})
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue(), "Namespace should be deleted")

			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, &securityv1alpha1.Organization{})
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue(), "Organization should be deleted")
		})
	})
})
