package namespace

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("org-%s", org.Name),
			Labels: map[string]string{
				"giantswarm.io/organization": org.Name,
			},
		},
	}

	err := r.client.Create(ctx, namespace)
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (r *Resource) EnsureDeleted(ctx context.Context, org *securityv1alpha1.Organization) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("org-%s", org.Name),
		},
	}

	err := r.client.Delete(ctx, namespace)
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}
