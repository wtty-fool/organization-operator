package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
	organization "github.com/giantswarm/organization-operator/pkg/namespace"
)

const organizationFinalizerName = "security.giantswarm.io/organization-finalizer"

// OrganizationReconciler reconciles a Organization object
type OrganizationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// nolint: lll
//+kubebuilder:rbac:groups=security.giantswarm.io,resources=organizations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.giantswarm.io,resources=organizations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=security.giantswarm.io,resources=organizations/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

func (r *OrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("organization", req.NamespacedName)
	log.Info("Starting reconciliation")

	orgResource, err := organization.New(organization.Config{
		K8sClient: r.Client,
		Logger:    log,
	})
	if err != nil {
		log.Error(err, "Failed to create organization resource")
		return ctrl.Result{}, fmt.Errorf("failed to create organization resource: %w", err)
	}

	var org securityv1alpha1.Organization
	if err := r.Get(ctx, req.NamespacedName, &org); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, it could have been deleted after reconcile request.
			// We need to handle the case where the namespace still exists.
			log.Info("Organization resource not found. Checking for orphaned namespace")
			return ctrl.Result{}, orgResource.EnsureDeleted(ctx, &securityv1alpha1.Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name: req.Name,
				},
			})
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Organization")
		return ctrl.Result{}, err
	}

	// Check if the Organization instance is marked to be deleted
	if org.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(&org, organizationFinalizerName) {
			// Run finalization logic. If it fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := orgResource.EnsureDeleted(ctx, &org); err != nil {
				log.Error(err, "Failed to finalize organization")
				return ctrl.Result{}, err
			}

			// Remove the finalizer. Once all finalizers have been removed, the object will be deleted.
			controllerutil.RemoveFinalizer(&org, organizationFinalizerName)
			if err := r.Update(ctx, &org); err != nil {
				log.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR if it doesn't exist
	if !controllerutil.ContainsFinalizer(&org, organizationFinalizerName) {
		controllerutil.AddFinalizer(&org, organizationFinalizerName)
		if err := r.Update(ctx, &org); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Ensure the organization is created
	log.Info("Ensuring organization creation")
	if err := orgResource.EnsureCreated(ctx, &org); err != nil {
		log.Error(err, "Failed to ensure organization creation")
		return ctrl.Result{}, fmt.Errorf("failed to ensure organization creation: %w", err)
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
