package controller

import (
	securityv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/organization-operator/pkg/project"
	organization "github.com/giantswarm/organization-operator/service/controller/resource/organization"
)

type OrganizationConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
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
