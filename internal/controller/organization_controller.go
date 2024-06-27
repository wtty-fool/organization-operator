package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
)

const organizationFinalizerName = "organization.giantswarm.io/finalizer"

type OrganizationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *OrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var organization securityv1alpha1.Organization
	if err := r.Get(ctx, req.NamespacedName, &organization); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Organization resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Organization")
		return ctrl.Result{}, err
	}

	// Check if the organization is marked to be deleted
	if !organization.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, &organization)
	}

	// Add finalizer if it doesn't exist
	if !hasOrganizationFinalizer(&organization) {
		organization.ObjectMeta.Finalizers = append(organization.ObjectMeta.Finalizers, organizationFinalizerName)
		if err := r.Update(ctx, &organization); err != nil {
			logger.Error(err, "Failed to update Organization with finalizer")
			return ctrl.Result{}, err
		}
	}

	// Create or update the namespace
	namespaceName := fmt.Sprintf("org-%s", organization.Name)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				"giantswarm.io/organization": organization.Name,
			},
		},
	}

	if err := r.Create(ctx, namespace); err != nil {
		if !errors.IsAlreadyExists(err) {
			logger.Error(err, "Failed to create namespace")
			return ctrl.Result{}, err
		}
	}

	// Update status
	organization.Status.Namespace = namespaceName
	if err := r.Status().Update(ctx, &organization); err != nil {
		logger.Error(err, "Failed to update Organization status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *OrganizationReconciler) reconcileDelete(ctx context.Context, organization *securityv1alpha1.Organization) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Delete the associated namespace
	namespaceName := fmt.Sprintf("org-%s", organization.Name)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	if err := r.Delete(ctx, namespace); err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to delete namespace")
		return ctrl.Result{}, err
	}

	// Remove the finalizer
	organization.ObjectMeta.Finalizers = removeOrganizationFinalizer(organization.ObjectMeta.Finalizers)
	if err := r.Update(ctx, organization); err != nil {
		logger.Error(err, "Failed to remove finalizer")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully deleted organization and removed finalizer")
	return ctrl.Result{}, nil
}

func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.Organization{}).
		Complete(r)
}

func hasOrganizationFinalizer(organization *securityv1alpha1.Organization) bool {
	for _, finalizer := range organization.ObjectMeta.Finalizers {
		if finalizer == organizationFinalizerName {
			return true
		}
	}
	return false
}

func removeOrganizationFinalizer(finalizers []string) []string {
	result := []string{}
	for _, finalizer := range finalizers {
		if finalizer != organizationFinalizerName {
			result = append(result, finalizer)
		}
	}
	return result
}
