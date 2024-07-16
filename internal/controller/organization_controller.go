/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
	"github.com/giantswarm/organization-operator/internal/util"
)

const finalizerName = "security.giantswarm.io/organization-finalizer"

// OrganizationReconciler reconciles a Organization object
type OrganizationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *OrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Starting reconciliation", "organization", req.NamespacedName)

	organization := &securityv1alpha1.Organization{}
	if err := r.Get(ctx, req.NamespacedName, organization); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Organization resource not found. Ignoring since object must be deleted", "organization", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Organization", "organization", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if !organization.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("Organization is being deleted", "organization", req.NamespacedName)
		return r.reconcileDelete(ctx, organization)
	}

	if !controllerutil.ContainsFinalizer(organization, finalizerName) {
		log.Info("Adding finalizer to Organization", "organization", req.NamespacedName, "finalizer", finalizerName)
		controllerutil.AddFinalizer(organization, finalizerName)
		if err := r.Update(ctx, organization); err != nil {
			log.Error(err, "Failed to update Organization with finalizer", "organization", req.NamespacedName)
			return ctrl.Result{}, err
		}
	}

	namespace := util.CreateNamespace(organization.Name)

	if err := util.EnsureNamespace(ctx, r.Client, namespace); err != nil {
		log.Error(err, "Failed to ensure Namespace", "namespace", namespace.Name)
		return ctrl.Result{}, err
	}

	if organization.Status.Namespace != namespace.Name {
		log.Info("Updating Organization status with namespace", "organization", req.NamespacedName, "namespace", namespace.Name)
		organization.Status.Namespace = namespace.Name
		if err := r.Status().Update(ctx, organization); err != nil {
			log.Error(err, "Failed to update Organization status", "organization", req.NamespacedName)
			return ctrl.Result{}, err
		}
	}

	log.Info("Reconciliation completed successfully", "organization", req.NamespacedName)
	return ctrl.Result{}, nil
}

func (r *OrganizationReconciler) reconcileDelete(ctx context.Context, organization *securityv1alpha1.Organization) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Reconciling deletion", "organization", organization.Name)

	if err := util.DeleteNamespace(ctx, r.Client, organization.Name); err != nil {
		log.Error(err, "Failed to delete Namespace", "organization", organization.Name)
		return ctrl.Result{}, err
	}

	if controllerutil.ContainsFinalizer(organization, finalizerName) {
		log.Info("Removing finalizer from Organization", "organization", organization.Name, "finalizer", finalizerName)
		controllerutil.RemoveFinalizer(organization, finalizerName)
		if err := r.Update(ctx, organization); err != nil {
			log.Error(err, "Failed to remove finalizer from Organization", "organization", organization.Name)
			return ctrl.Result{}, err
		}
	}

	log.Info("Deletion reconciliation completed successfully", "organization", organization.Name)
	return ctrl.Result{}, nil
}

func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.Organization{}).
		Complete(r)
}
