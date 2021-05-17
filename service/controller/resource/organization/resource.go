package organization

import (
	"fmt"

	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	companyclient "github.com/giantswarm/companyd-client-go"

	"github.com/giantswarm/organization-operator/pkg/label"
	"github.com/giantswarm/organization-operator/pkg/project"
)

const (
	Name                   = "organization"
	organizationNamePrefix = "org-"
)

var (
	forbiddenOrganizationPrefixes = []string{
		"default",
		"kube-",
		"monitoring",
		"gatekeeper",
		"draughtsman",
	}
)

type Config struct {
	K8sClient       k8sclient.Interface
	Logger          micrologger.Logger
	LegacyOrgClient *companyclient.Client
}

type Resource struct {
	k8sClient       k8sclient.Interface
	logger          micrologger.Logger
	legacyOrgClient *companyclient.Client
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {

		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.LegacyOrgClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.LegacyOrgClient must not be empty", config)
	}

	r := &Resource{
		k8sClient:       config.K8sClient,
		logger:          config.Logger,
		legacyOrgClient: config.LegacyOrgClient,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func newOrganizationNamespace(organizationName string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s%s", organizationNamePrefix, organizationName),
			Labels: map[string]string{
				label.Organization: organizationName,
				label.ManagedBy:    project.Name(),
			},
		},
	}

}
