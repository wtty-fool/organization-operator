package organization

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/organization-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	org, err := key.ToOrganization(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.k8sClient.K8sClient().CoreV1().Namespaces().Delete(
		fmt.Sprintf("%s%s", organizationNamePrefix, org.ObjectMeta.Name),
		&metav1.DeleteOptions{},
	)
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("organization namespace %#q does not exist", org.ObjectMeta.Name))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("organization namespace %#q has been deleted", org.ObjectMeta.Name))
	return nil
}
