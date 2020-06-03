package organization

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/organization-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	org, err := key.ToOrganization(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	orgNamespace := newOrganizationNamespace(org.Name)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting organization namespace %#q", orgNamespace.Name))

	err = r.k8sClient.CtrlClient().Delete(context.Background(), orgNamespace)
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("organization namespace %#q does not exist", orgNamespace.Name))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted organization namespace %#q", orgNamespace.Name))
	return nil
}
