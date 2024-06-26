package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
	"github.com/giantswarm/organization-operator/pkg/resources/namespace"
)

// OrganizationReconciler reconciles a Organization object
type OrganizationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *OrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Organization", "namespacedName", req.NamespacedName)

	var organization securityv1alpha1.Organization
	err := r.Get(ctx, req.NamespacedName, &organization)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Organization not found. Attempting to delete associated namespace.")
			return r.deleteAssociatedNamespace(ctx, req.Name)
		}
		logger.Error(err, "Unable to fetch Organization")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	namespaceResource := namespace.New(r.Client)

	if organization.DeletionTimestamp.IsZero() {
		logger.Info("Organization is not being deleted, ensuring namespace exists")
		if err := namespaceResource.EnsureCreated(ctx, &organization); err != nil {
			logger.Error(err, "Failed to create namespace")
			return ctrl.Result{}, err
		}

		if organization.Status.Namespace == "" {
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				// Fetch the latest state of the organization to minimize the chance of a conflict error
				if getErr := r.Get(ctx, req.NamespacedName, &organization); getErr != nil {
					return getErr
				}
				organization.Status.Namespace = fmt.Sprintf("org-%s", organization.Name)
				// Update operation
				return r.Status().Update(ctx, &organization)
			})
			if err != nil {
				logger.Error(err, "Failed to update Organization status after retries")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil // Successfully updated
		}
	} else {
		logger.Info("Organization is being deleted, ensuring namespace is deleted")
		if err := namespaceResource.EnsureDeleted(ctx, &organization); err != nil {
			logger.Error(err, "Failed to delete namespace")
			return ctrl.Result{}, err
		}
		logger.Info("Namespace deleted successfully")
	}

	logger.Info("Reconciliation completed successfully")
	return ctrl.Result{}, nil
}

func (r *OrganizationReconciler) deleteAssociatedNamespace(ctx context.Context, organizationName string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	namespaceName := fmt.Sprintf("org-%s", organizationName)

	namespaceResource := namespace.New(r.Client)
	dummyOrg := &securityv1alpha1.Organization{
		Spec: securityv1alpha1.OrganizationSpec{
			Name: organizationName,
		},
	}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return namespaceResource.EnsureDeleted(ctx, dummyOrg)
	})
	if err != nil {
		logger.Error(err, "Failed to delete namespace after retries", "namespace", namespaceName)
		return ctrl.Result{}, err
	}
	logger.Info("Associated namespace deleted successfully", "namespace", namespaceName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.Organization{}).
		Complete(r)
}
