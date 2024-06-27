package organization

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/finalizerskeptcontext"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/organization-operator/controllers/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	org, err := key.ToOrganization(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	orgNamespace := newOrganizationNamespace(org.Name)

	err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Name: orgNamespace.Name}, orgNamespace)
	if err == nil {
		finalizerskeptcontext.SetKept(ctx)
		if orgNamespace.DeletionTimestamp != nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for deletion of organization namespace %#q", orgNamespace.Name))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting organization namespace %#q", orgNamespace.Name))
			err = r.k8sClient.CtrlClient().Delete(context.Background(), orgNamespace)
		}
	}

	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("organization namespace %#q does not exist", orgNamespace.Name))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted organization namespace %#q", orgNamespace.Name))
	return nil
}
