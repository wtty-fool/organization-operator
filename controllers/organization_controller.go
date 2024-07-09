package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
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

func (r *OrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	start := time.Now()
	defer func() {
		reconcileDuration.Observe(time.Since(start).Seconds())
	}()

	log := r.Log.WithValues("organization", req.NamespacedName)
	log.Info("Starting reconciliation")

	orgResource, err := organization.New(organization.Config{
		K8sClient: r.Client,
		Logger:    log,
	})
	if err != nil {
		log.Error(err, "Failed to create organization resource")
		reconcileErrors.Inc()
		return ctrl.Result{}, fmt.Errorf("failed to create organization resource: %w", err)
	}

	var org securityv1alpha1.Organization
	if err := r.Get(ctx, req.NamespacedName, &org); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, it could have been deleted after reconcile request.
			log.Info("Organization resource not found. Checking for orphaned namespace")
			err = orgResource.EnsureDeleted(ctx, &securityv1alpha1.Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name: req.Name,
				},
			})
			if err == nil {
				r.updateOrganizationCount(ctx)
				namespacesExist.DeleteLabelValues(req.Name)
			}
			return ctrl.Result{}, err
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Organization")
		reconcileErrors.Inc()
		return ctrl.Result{}, err
	}

	// Check if the Organization instance is marked to be deleted
	if org.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(&org, organizationFinalizerName) {
			// Run finalization logic. If it fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := orgResource.EnsureDeleted(ctx, &org); err != nil {
				log.Error(err, "Failed to finalize organization")
				reconcileErrors.Inc()
				return ctrl.Result{}, err
			}

			// Remove the finalizer. Once all finalizers have been removed, the object will be deleted.
			controllerutil.RemoveFinalizer(&org, organizationFinalizerName)
			if err := r.Update(ctx, &org); err != nil {
				log.Error(err, "Failed to remove finalizer")
				reconcileErrors.Inc()
				return ctrl.Result{}, err
			}
			r.updateOrganizationCount(ctx)
			namespacesExist.DeleteLabelValues(org.Name)
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR if it doesn't exist
	if !controllerutil.ContainsFinalizer(&org, organizationFinalizerName) {
		controllerutil.AddFinalizer(&org, organizationFinalizerName)
		if err := r.Update(ctx, &org); err != nil {
			log.Error(err, "Failed to add finalizer")
			reconcileErrors.Inc()
			return ctrl.Result{}, err
		}
	}

	// Ensure the organization is created
	log.Info("Ensuring organization creation")
	if err := orgResource.EnsureCreated(ctx, &org); err != nil {
		log.Error(err, "Failed to ensure organization creation")
		reconcileErrors.Inc()
		return ctrl.Result{}, fmt.Errorf("failed to ensure organization creation: %w", err)
	}
	r.updateOrganizationCount(ctx)

	// Check if the namespace exists
	namespace := &corev1.Namespace{}
	err = r.Get(ctx, client.ObjectKey{Name: org.Status.Namespace}, namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			namespacesExist.WithLabelValues(org.Name).Set(0)
		} else {
			log.Error(err, "Failed to get namespace")
			reconcileErrors.Inc()
		}
	} else {
		namespacesExist.WithLabelValues(org.Name).Set(1)
	}

	log.Info("Reconciliation completed successfully")
	return ctrl.Result{}, nil
}

func (r *OrganizationReconciler) updateOrganizationCount(ctx context.Context) {
	var orgList securityv1alpha1.OrganizationList
	if err := r.List(ctx, &orgList); err != nil {
		r.Log.Error(err, "Failed to list organizations")
		return
	}
	organizationCount.Set(float64(len(orgList.Items)))
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.Organization{}).
		Complete(r)
}
