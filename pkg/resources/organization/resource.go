package organization

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/organization-operator/pkg/label"
	"github.com/giantswarm/organization-operator/pkg/project"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
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
	K8sClient client.Client
	Logger    logr.Logger
}

type Resource struct {
	k8sClient client.Client
	logger    logr.Logger
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var emptyLogger logr.Logger
	if config.Logger == emptyLogger {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	r.logger.Info("Organization resource created")

	return r, nil
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
