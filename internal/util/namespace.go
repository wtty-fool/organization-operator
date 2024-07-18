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
package util

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func CreateNamespace(orgName string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "org-" + orgName,
			Labels: map[string]string{
				"giantswarm.io/organization": orgName,
				"giantswarm.io/managed-by":   "organization-operator",
			},
		},
	}
}

// EnsureNamespace creates a namespace if it doesn't exist
func EnsureNamespace(ctx context.Context, c client.Client, namespace *corev1.Namespace) error {
	existingNamespace := &corev1.Namespace{}
	err := c.Get(ctx, client.ObjectKey{Name: namespace.Name}, existingNamespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return c.Create(ctx, namespace)
		}
		return err
	}

	// Update labels if they don't match
	if !reflect.DeepEqual(existingNamespace.Labels, namespace.Labels) {
		existingNamespace.Labels = namespace.Labels
		return c.Update(ctx, existingNamespace)
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
