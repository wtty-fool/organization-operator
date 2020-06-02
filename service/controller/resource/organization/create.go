package organization

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/organization-operator/pkg/label"
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

	newNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s%s", organizationNamePrefix, org.ObjectMeta.Name),
			Labels: map[string]string{
				label.Organization: org.ObjectMeta.Name,
			},
		},
	}

	_, err = r.k8sClient.K8sClient().CoreV1().Namespaces().Create(&newNamespace)
	if apierrors.IsAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("organization namespace %#q already exists", newNamespace.ObjectMeta.Name))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("organization namespace %#q has been created", newNamespace.ObjectMeta.Name))
	return nil
}
