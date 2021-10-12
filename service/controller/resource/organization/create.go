package organization

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	companyclient "github.com/giantswarm/companyd-client-go"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/organization-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	org, err := key.ToOrganization(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, prefix := range forbiddenOrganizationPrefixes {
		if strings.HasPrefix(org.Name, prefix) {
			r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("organization name %#q cannot start with %q", org.Name, prefix))
			return nil
		}
	}

	err = r.ensureOrganizationHasSubscriptionIdAnnotation(ctx, org)
	if err != nil {
		return microerror.Mask(err)
	}

	orgNamespace := newOrganizationNamespace(org.Name)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating organization namespace %#q", orgNamespace.Name))

	err = r.k8sClient.CtrlClient().Create(ctx, orgNamespace)
	if apierrors.IsAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("organization namespace %#q already exists", orgNamespace.Name))
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created organization namespace %#q", orgNamespace.Name))

	org.Status.Namespace = orgNamespace.Name
	err = r.k8sClient.CtrlClient().Status().Update(ctx, &org)
	if err != nil {
		return microerror.Mask(err)
	}

	legacyOrgName := key.LegacyOrganizationName(&org)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating legacy organization %#q", legacyOrgName))

	legacyOrgFields := companyclient.CompanyFields{
		DefaultCluster: "deprecated",
	}
	err = r.legacyOrgClient.CreateCompany(legacyOrgName, legacyOrgFields)
	if companyclient.IsErrCompanyAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("legacy organization %#q already exists", legacyOrgName))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created legacy organization %#q", legacyOrgName))

	return nil
}

func (r *Resource) ensureOrganizationHasSubscriptionIdAnnotation(ctx context.Context, organization v1alpha1.Organization) error {
	// Retrieve secret related to this organization.
	secret, err := findSecret(ctx, r.k8sClient.CtrlClient(), organization.Name)
	if IsSecretNotFound(err) {
		// We don't want this error to block execution so we still return nil and just log the problem.
		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("unable to find a secret for organization %s. Cannot set subscriptionid annotation", organization.Name))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	// The subscription id field is missing in non azure installations so it's ok.
	if subscription, ok := secret.Data["azure.azureoperator.subscriptionid"]; ok {
		organization.Annotations["subscription"] = string(subscription)
		err = r.k8sClient.CtrlClient().Update(ctx, &organization)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func findSecret(ctx context.Context, client ctrl.Client, orgName string) (*corev1.Secret, error) {
	// Look for a secret with labels "app: credentiald" and "giantswarm.io/organization: org"
	secrets := &corev1.SecretList{}

	err := client.List(ctx, secrets, ctrl.MatchingLabels{"app": "credentiald", label.Organization: orgName})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(secrets.Items) > 0 {
		return &secrets.Items[0], nil
	}
	secret := &corev1.Secret{}

	// Organization-specific secret not found, use secret named "credential-default".
	err = client.Get(ctx, ctrl.ObjectKey{Namespace: "giantswarm", Name: "credential-default"}, secret)
	if apierrors.IsNotFound(err) {
		return nil, microerror.Maskf(secretNotFoundError, "Unable to find secret for organization %s", orgName)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	return secret, nil
}
