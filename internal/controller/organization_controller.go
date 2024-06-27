package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
	"github.com/giantswarm/organization-operator/pkg/resources/namespace"
)

type OrganizationReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	ResyncPeriod time.Duration
}

func (r *OrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var organization corev1alpha1.Organization
	if err := r.Get(ctx, req.NamespacedName, &organization); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Organization resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Organization")
		return ctrl.Result{}, err
	}

	namespaceResource := namespace.New(r.Client)

	if organization.DeletionTimestamp.IsZero() {
		if err := namespaceResource.EnsureCreated(ctx, &organization); err != nil {
			logger.Error(err, "Failed to create namespace")
			return ctrl.Result{}, err
		}
	} else {
		if err := namespaceResource.EnsureDeleted(ctx, &organization); err != nil {
			logger.Error(err, "Failed to delete namespace")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: r.ResyncPeriod}, nil
}

func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Organization{}).
		Complete(r)
}
