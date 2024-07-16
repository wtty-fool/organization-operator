package util

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func CreateNamespace(orgName string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("org-%s", orgName),
		},
	}
}

// EnsureNamespace creates a namespace if it doesn't exist
func EnsureNamespace(ctx context.Context, c client.Client, namespace *corev1.Namespace) error {
	log := log.FromContext(ctx)

	err := c.Get(ctx, client.ObjectKey{Name: namespace.Name}, &corev1.Namespace{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Namespace not found, creating", "namespace", namespace.Name)
			err = c.Create(ctx, namespace)
			if err != nil {
				log.Error(err, "Failed to create namespace", "namespace", namespace.Name)
				return fmt.Errorf("failed to create namespace %s: %w", namespace.Name, err)
			}
			log.Info("Successfully created namespace", "namespace", namespace.Name)
		} else {
			log.Error(err, "Failed to get namespace", "namespace", namespace.Name)
			return fmt.Errorf("failed to get namespace %s: %w", namespace.Name, err)
		}
	} else {
		log.Info("Namespace already exists", "namespace", namespace.Name)
	}
	return nil
}

// DeleteNamespace deletes the namespace for the given organization name
func DeleteNamespace(ctx context.Context, c client.Client, orgName string) error {
	log := log.FromContext(ctx)
	namespace := CreateNamespace(orgName)

	log.Info("Attempting to delete namespace", "namespace", namespace.Name)
	err := c.Delete(ctx, namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Namespace not found, skipping deletion", "namespace", namespace.Name)
			return nil
		}
		log.Error(err, "Failed to delete namespace", "namespace", namespace.Name)
		return fmt.Errorf("failed to delete namespace %s: %w", namespace.Name, err)
	}
	log.Info("Successfully deleted namespace", "namespace", namespace.Name)
	return nil
}
