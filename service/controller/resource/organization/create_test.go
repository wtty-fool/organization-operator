package organization

import (
	"context"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/organization-operator/api/v1alpha1"
	"github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mock_organization "github.com/giantswarm/organization-operator/service/controller/resource/organization/mock_spec"

	"github.com/giantswarm/organization-operator/service/unittest"
)

func Test_NamespaceIsCreated(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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

	companydClientMock := mock_organization.NewMockCompanydClient(ctrl)
	companydClientMock.
		EXPECT().
		CreateCompany(org.Name, gomock.Any()).
		Return(nil)

	credentialdClientMock := mock_organization.NewMockCredentialdClient(ctrl)

	config := Config{
		K8sClient:              k8sClient,
		Logger:                 logger,
		LegacyOrgClient:        companydClientMock,
		LegacyCredentialClient: credentialdClientMock,
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
