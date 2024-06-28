package controllers

import (
	"context"
	"fmt"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
	"github.com/giantswarm/organization-operator/pkg/resources/organization"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OrganizationReconciler reconciles a Organization object
type OrganizationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// nolint:lll
//+kubebuilder:rbac:groups=security.giantswarm.io,resources=organizations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.giantswarm.io,resources=organizations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=security.giantswarm.io,resources=organizations/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

func (r *OrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("organization", req.NamespacedName)
	log.Info("Starting reconciliation")

	var org securityv1alpha1.Organization
	if err := r.Get(ctx, req.NamespacedName, &org); err != nil {
		log.Error(err, "unable to fetch Organization")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Creating organization resource")
	orgResource, err := organization.New(organization.Config{
		K8sClient: r.Client,
		Logger:    log,
	})
	if err != nil {
		log.Error(err, "failed to create organization resource")
		return ctrl.Result{}, fmt.Errorf("failed to create organization resource: %w", err)
	}

	if org.DeletionTimestamp.IsZero() {
		log.Info("Ensuring organization creation")
		if err := orgResource.EnsureCreated(ctx, &org); err != nil {
			log.Error(err, "failed to ensure organization creation")
			return ctrl.Result{}, fmt.Errorf("failed to ensure organization creation: %w", err)
		}
	} else {
		log.Info("Ensuring organization deletion")
		if err := orgResource.EnsureDeleted(ctx, &org); err != nil {
			log.Error(err, "failed to ensure organization deletion")
			return ctrl.Result{}, fmt.Errorf("failed to ensure organization deletion: %w", err)
		}
	}

	log.Info("Reconciliation completed successfully")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.Organization{}).
		Complete(r)
}
