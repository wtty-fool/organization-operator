package organization

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/organization-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	org, err := key.ToOrganization(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, prefix := range forbiddenOrganizationPrefixes {
		if strings.HasPrefix(org.ObjectMeta.Name, prefix) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("organization name %#q cannot start with %q", org.ObjectMeta.Name, prefix))
			return nil
		}
	}

	orgNamespace := newOrganizationNamespace(org.ObjectMeta.Name)

	err = r.k8sClient.CtrlClient().Create(context.Background(), orgNamespace)
	if apierrors.IsAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("organization namespace %#q already exists", orgNamespace.ObjectMeta.Name))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("organization namespace %#q has been created", orgNamespace.ObjectMeta.Name))
	return nil
}
