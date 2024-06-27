package organization

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/organization-operator/api/v1alpha1"
	"github.com/giantswarm/organization-operator/controllers/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	org, err := key.ToOrganization(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, prefix := range forbiddenOrganizationPrefixes {
		if strings.HasPrefix(org.Name, prefix) {
			r.logger.Info(fmt.Sprintf("organization name %#q cannot start with %q", org.Name, prefix))
			return nil
		}
	}

	err = r.ensureOrganizationHasSubscriptionIdAnnotation(ctx, org)
	if err != nil {
		return microerror.Mask(err)
	}

	orgNamespace := newOrganizationNamespace(org.Name)
	r.logger.Info(fmt.Sprintf("creating organization namespace %#q", orgNamespace.Name))

	err = r.k8sClient.Create(ctx, orgNamespace)
	if apierrors.IsAlreadyExists(err) {
		r.logger.Info(fmt.Sprintf("organization namespace %#q already exists", orgNamespace.Name))
		err := r.ensureOrganizationNamespaceHasOrganizationLabels(ctx, orgNamespace)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Info(fmt.Sprintf("created organization namespace %#q", orgNamespace.Name))

	patch := []byte(fmt.Sprintf(`{"status":{"namespace": "%s"}}`, orgNamespace.Name))
	err = r.k8sClient.Status().Patch(ctx, &org, ctrl.RawPatch(types.MergePatchType, patch))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) ensureOrganizationHasSubscriptionIdAnnotation(ctx context.Context,
	organization v1alpha1.Organization) error {
	r.logger.Info(fmt.Sprintf("ensuring organization %q has subscriptionid annotation", organization.Name))
	// Retrieve secret related to this organization.
	secret, err := findSecret(ctx, r.k8sClient, organization.Name)
	if IsSecretNotFound(err) {
		// We don't want this error to block execution so we still return nil and just log the problem.
		r.logger.Info(fmt.Sprintf("unable to find a secret for organization %s. Cannot set subscriptionid annotation",
			organization.Name))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	// The subscription id field is missing in non azure installations so it's ok.
	if subscription, ok := secret.Data["azure.azureoperator.subscriptionid"]; ok && len(subscription) > 0 {
		r.logger.Info(fmt.Sprintf("setting subscriptionid annotation to %q for organization %q",
			string(subscription), organization.Name))
		patch := []byte(fmt.Sprintf(`{"metadata":{"annotations":{"subscription": "%s"}}}`, string(subscription)))
		err = r.k8sClient.Patch(ctx, &organization, ctrl.RawPatch(types.MergePatchType, patch))
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		r.logger.Info(fmt.Sprintf("azure.azureoperator.subscriptionid field not found or empty in secret %q", secret.Name))
	}

	return nil
}

func (r *Resource) ensureOrganizationNamespaceHasOrganizationLabels(ctx context.Context,
	namespace *corev1.Namespace) error {
	r.logger.Info(fmt.Sprintf("ensuring organization namespace %#q has organization labels", namespace.Name))

	currentNamespace := &corev1.Namespace{}
	err := r.k8sClient.Get(ctx, ctrl.ObjectKey{Name: namespace.Name}, currentNamespace)
	if err != nil {
		return microerror.Mask(err)
	}
	for key, value := range namespace.Labels {
		if currentNamespace.Labels[key] != value {
			r.logger.Info(fmt.Sprintf("namespace %#q has label %q=%q, but should have %q=%q",
				namespace.Name, key, currentNamespace.Labels[key], key, value))
			patch := []byte(fmt.Sprintf(`{"metadata":{"labels":{"%s": "%s"}}}`, key, value))
			err = r.k8sClient.Patch(ctx, namespace, ctrl.RawPatch(types.MergePatchType, patch))
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	r.logger.Info(fmt.Sprintf("ensured organization namespace %#q has organization labels", namespace.Name))

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
