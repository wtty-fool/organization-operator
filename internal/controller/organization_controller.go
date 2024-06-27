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

	"github.com/giantswarm/organization-operator/api/v1alpha1"
	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
)

const (
	organizationFinalizerName = "organization.giantswarm.io/finalizer"
)

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
	if !containsString(organization.ObjectMeta.Finalizers, organizationFinalizerName) {
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

func (r *OrganizationReconciler) reconcileDelete(ctx context.Context, organization *v1alpha1.Organization) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	namespaceName := fmt.Sprintf("org-%s", organization.Name)

	namespace := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: namespaceName}, namespace); err == nil {
		logger.Info("Attempting to delete namespace", "namespace", namespaceName, "finalizers", namespace.Finalizers)
		if err := r.Delete(ctx, namespace); err != nil {
			logger.Error(err, "Failed to delete namespace", "namespace", namespaceName)
			return ctrl.Result{}, err
		}
	} else if !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get namespace for deletion", "namespace", namespaceName)
		return ctrl.Result{}, err
	}

	if containsString(organization.ObjectMeta.Finalizers, organizationFinalizerName) {
		organization.ObjectMeta.Finalizers = removeString(organization.ObjectMeta.Finalizers, organizationFinalizerName)
		if err := r.Update(ctx, organization); err != nil {
			logger.Error(err, "Failed to update Organization during finalizer removal", "organization", organization.Name)
			return ctrl.Result{}, err
		}
		logger.Info("Finalizer removed", "organization", organization.Name)
	}

	logger.Info("Successfully deleted organization and namespace", "namespace", namespaceName)
	return ctrl.Result{}, nil
}

func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.Organization{}).
		Complete(r)
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	result := []string{}
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
