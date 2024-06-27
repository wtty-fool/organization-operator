package namespace

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
)

type Resource struct {
	client client.Client
}

func New(client client.Client) *Resource {
	return &Resource{
		client: client,
	}
}

func (r *Resource) EnsureCreated(ctx context.Context, org *securityv1alpha1.Organization) error {
	logger := log.FromContext(ctx)
	namespaceName := fmt.Sprintf("org-%s", org.Name)

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				"giantswarm.io/organization": org.Name,
			},
		},
	}

	logger.Info("Creating namespace", "namespace", namespaceName)
	err := r.client.Create(ctx, namespace)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			logger.Info("Namespace already exists", "namespace", namespaceName)
			return r.updateNamespaceIfNeeded(ctx, namespace)
		}
		logger.Error(err, "Failed to create namespace", "namespace", namespaceName)
		return err
	}

	logger.Info("Namespace created successfully", "namespace", namespaceName)
	return nil
}

func (r *Resource) EnsureDeleted(ctx context.Context, org *securityv1alpha1.Organization) error {
	logger := log.FromContext(ctx)
	namespaceName := fmt.Sprintf("org-%s", org.Name)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	logger.Info("Deleting namespace", "namespace", namespaceName)
	err := r.client.Delete(ctx, namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Namespace not found, already deleted", "namespace", namespaceName)
			return nil
		}
		logger.Error(err, "Failed to delete namespace", "namespace", namespaceName)
		return err
	}

	logger.Info("Namespace deleted successfully", "namespace", namespaceName)
	return nil
}

func (r *Resource) updateNamespaceIfNeeded(ctx context.Context, namespace *corev1.Namespace) error {
	logger := log.FromContext(ctx)
	currentNamespace := &corev1.Namespace{}
	err := r.client.Get(ctx, client.ObjectKey{Name: namespace.Name}, currentNamespace)
	if err != nil {
		return err
	}

	if !namespacesEqual(currentNamespace, namespace) {
		logger.Info("Updating namespace", "namespace", namespace.Name)
		err = r.client.Update(ctx, namespace)
		if err != nil {
			logger.Error(err, "Failed to update namespace", "namespace", namespace.Name)
			return err
		}
		logger.Info("Namespace updated successfully", "namespace", namespace.Name)
	}
	return nil
}

func namespacesEqual(a, b *corev1.Namespace) bool {
	return a.Labels["giantswarm.io/organization"] == b.Labels["giantswarm.io/organization"]
}
