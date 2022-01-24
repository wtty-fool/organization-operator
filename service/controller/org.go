package controller

import (
	securityv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v5/pkg/controller"
	"github.com/giantswarm/operatorkit/v5/pkg/resource"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/crud"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/k8s/configmapresource"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/wrapper/retryresource"
	"k8s.io/apimachinery/pkg/runtime"

	companyclient "github.com/giantswarm/companyd-client-go"
	credentialclient "github.com/giantswarm/credentiald/v2/client"

	"github.com/giantswarm/organization-operator/pkg/project"
	"github.com/giantswarm/organization-operator/service/controller/resource/organization"
	"github.com/giantswarm/organization-operator/service/controller/resource/orgconfigmap"
)

type OrganizationConfig struct {
	BaseDomain             string
	K8sClient              k8sclient.Interface
	Logger                 micrologger.Logger
	LegacyOrgClient        *companyclient.Client
	LegacyCredentialClient *credentialclient.Client
}

type Organization struct {
	*controller.Controller
}

func NewOrganization(config OrganizationConfig) (*Organization, error) {
	var err error

	resources, err := newOrganizationResources(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(securityv1alpha1.Organization)
			},
			Resources: resources,

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/organization-operator-todo-controller.
			Name: project.Name() + "-organization-controller",
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Organization{
		Controller: operatorkitController,
	}

	return c, nil
}

func newOrganizationResources(config OrganizationConfig) ([]resource.Interface, error) {
	var err error

	var orgResource resource.Interface
	{
		c := organization.Config{
			K8sClient:              config.K8sClient,
			Logger:                 config.Logger,
			LegacyOrgClient:        config.LegacyOrgClient,
			LegacyCredentialClient: config.LegacyCredentialClient,
		}

		orgResource, err = organization.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterConfigMapGetter configmapresource.StateGetter
	{
		c := orgconfigmap.Config{
			BaseDomain: config.BaseDomain,
			K8sClient:  config.K8sClient,
			Logger:     config.Logger,
		}

		clusterConfigMapGetter, err = orgconfigmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var orgConfigMapResource resource.Interface
	{
		c := configmapresource.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			AllowedLabels: []string{
				label.AppOperatorWatching,
			},
			Name:        orgconfigmap.Name,
			StateGetter: clusterConfigMapGetter,
		}

		ops, err := configmapresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		orgConfigMapResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		orgResource,
		orgConfigMapResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}

func toCRUDResource(logger micrologger.Logger, v crud.Interface) (*crud.Resource, error) {
	c := crud.ResourceConfig{
		CRUD:   v,
		Logger: logger,
	}

	r, err := crud.NewResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
