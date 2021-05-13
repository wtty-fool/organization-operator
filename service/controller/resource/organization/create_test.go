package organization

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/organization-operator/service/unittest"
)

func Test_NamespaceIsCreated(t *testing.T) {
	ctx := context.Background()
	logger := microloggertest.New()
	k8sClient := unittest.FakeK8sClient()
	ctrClient := k8sClient.CtrlClient()

	organizationName := "giantswarm"
	expectedNamespaceName := "org-giantswarm"

	org := &v1alpha1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: organizationName,
		},
		Spec: v1alpha1.OrganizationSpec{},
	}
	err := ctrClient.Create(ctx, org)
	if err != nil {
		t.Fatal(err)
	}

	config := Config{
		K8sClient: k8sClient,
		Logger:    logger,
	}
	organizationHandler, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	err = organizationHandler.EnsureCreated(ctx, org)
	if err != nil {
		t.Fatal(err)
	}

	orgNamespace := &v1.Namespace{}
	err = ctrClient.Get(ctx, client.ObjectKey{Name: expectedNamespaceName}, orgNamespace)
	if err != nil {
		t.Fatal(err)
	}

	err = ctrClient.Get(ctx, client.ObjectKey{Name: organizationName}, org)
	if err != nil {
		t.Fatal(err)
	}

	if org.Status.Namespace != expectedNamespaceName {
		t.Fatalf("created namespace should be stored in organization status")
	}
}
