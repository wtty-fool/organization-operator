package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
	"github.com/giantswarm/organization-operator/pkg/resources/organization"
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

func (r *OrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("organization", req.NamespacedName)

	var org securityv1alpha1.Organization
	if err := r.Get(ctx, req.NamespacedName, &org); err != nil {
		log.Error(err, "unable to fetch Organization")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	orgResource, err := organization.New(organization.Config{
		K8sClient: r.Client,
		Logger:    log,
	})
	if err != nil {
		log.Error(err, "failed to create organization resource")
		return ctrl.Result{}, err
	}

	if org.DeletionTimestamp.IsZero() {
		if err := orgResource.EnsureCreated(ctx, &org); err != nil {
			log.Error(err, "failed to ensure organization creation")
			return ctrl.Result{}, err
		}
	} else {
		if err := orgResource.EnsureDeleted(ctx, &org); err != nil {
			log.Error(err, "failed to ensure organization deletion")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.Organization{}).
		Complete(r)
}
