package organization

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/organization-operator/controllers/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	org, err := key.ToOrganization(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	nsName := fmt.Sprintf("org-%s", org.Name)
	ns := &corev1.Namespace{}
	err = r.k8sClient.Get(ctx, client.ObjectKey{Name: nsName}, ns)

	if apierrors.IsNotFound(err) {
		// The namespace is not found, which could mean it's an orphaned situation
		// or it's already been deleted. Let's handle potential orphaned namespace.
		return r.handleOrphanedNamespace(ctx, org.Name)
	} else if err != nil {
		// Error reading the namespace
		r.logger.Error(err, "Failed to get Namespace", "namespace", nsName)
		return microerror.Mask(err)
	}

	// Namespace exists and we found it
	if ns.DeletionTimestamp != nil {
		r.logger.Info(fmt.Sprintf("waiting for deletion of organization namespace %#q", nsName))
		return nil
	}

	r.logger.Info(fmt.Sprintf("deleting organization namespace %#q", nsName))
	err = r.k8sClient.Delete(ctx, ns)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Info(fmt.Sprintf("deleted organization namespace %#q", nsName))
	return nil
}

func (r *Resource) handleOrphanedNamespace(ctx context.Context, orgName string) error {
	nsName := fmt.Sprintf("org-%s", orgName)
	ns := &corev1.Namespace{}
	err := r.k8sClient.Get(ctx, client.ObjectKey{Name: nsName}, ns)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Namespace doesn't exist, nothing to do
			r.logger.Info("Namespace not found, no cleanup needed", "namespace", nsName)
			return nil
		}
		// Error reading the namespace
		r.logger.Error(err, "Failed to get Namespace", "namespace", nsName)
		return microerror.Mask(err)
	}

	// Namespace exists, we need to delete it
	r.logger.Info("Deleting orphaned namespace", "namespace", nsName)
	if err := r.k8sClient.Delete(ctx, ns); err != nil {
		r.logger.Error(err, "Failed to delete orphaned namespace", "namespace", nsName)
		return microerror.Mask(err)
	}

	return nil
}
