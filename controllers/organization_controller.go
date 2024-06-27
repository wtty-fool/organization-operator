package controllers

import (
	"time"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v7/pkg/controller"
	"github.com/giantswarm/operatorkit/v7/pkg/resource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/retryresource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
	"github.com/giantswarm/organization-operator/pkg/project"
	"github.com/giantswarm/organization-operator/pkg/resources/organization"
)

type OrganizationConfig struct {
	K8sClient    k8sclient.Interface
	Logger       micrologger.Logger
	ResyncPeriod *time.Duration
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
			NewRuntimeObjectFunc: func() client.Object {
				return new(securityv1alpha1.Organization)
			},
			Resources: resources,

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/organization-operator-todo-controller.
			Name: project.Name() + "-organization-controller",
		}

		if config.ResyncPeriod != nil {
			c.ResyncPeriod = *config.ResyncPeriod
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
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		orgResource, err = organization.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		orgResource,
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
