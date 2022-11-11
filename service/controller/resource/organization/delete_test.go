package organization

import (
	"context"
	"errors"
	"fmt"
	"github.com/giantswarm/credentiald/v2/service/lister"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/organization-operator/api/v1alpha1"
	mock_organization "github.com/giantswarm/organization-operator/service/controller/resource/organization/mock_spec"
	"github.com/giantswarm/organization-operator/service/unittest"
	"github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func Test_NamespaceIsDeleted(t *testing.T) {

	testCases := []struct {
		Name             string
		OrganizationName string
		CompanydError    error
		CredentialdError error
	}{
		{
			Name:             "case 0: Delete org with valid responses from companyd and credentiald",
			OrganizationName: "giantswarm",
			CompanydError:    nil,
			CredentialdError: nil,
		},
		{
			Name:             "case 1: Delete org with invalid response from companyd",
			OrganizationName: "giantswarm",
			CompanydError:    errors.New("NotFound"),
			CredentialdError: nil,
		},
		{
			Name:             "case 2: Delete org with invalid response from credentiald",
			OrganizationName: "giantswarm",
			CompanydError:    nil,
			CredentialdError: errors.New("NotFound"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			namespaceName := fmt.Sprintf("org-%s", tc.OrganizationName)

			org := &v1alpha1.Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name: tc.OrganizationName,
				},
				Spec: v1alpha1.OrganizationSpec{},
				Status: v1alpha1.OrganizationStatus{
					Namespace: namespaceName,
				},
			}

			namespace := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
				},
			}

			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			logger := microloggertest.New()
			k8sClient := unittest.FakeK8sClient(org, namespace)
			ctrClient := k8sClient.CtrlClient()

			orgNamespace := &v1.Namespace{}
			err := ctrClient.Get(ctx, client.ObjectKey{Name: namespaceName}, orgNamespace)
			if err != nil {
				t.Fatal(err)
			}

			credentialdClientMock := mock_organization.NewMockCredentialdClient(ctrl)
			credentialdClientMock.
				EXPECT().List(ctx, lister.Request{Organization: tc.OrganizationName}).
				Return([]lister.Response{}, tc.CredentialdError)

			companydClientMock := mock_organization.NewMockCompanydClient(ctrl)
			companydClientMock.
				EXPECT().
				DeleteCompany(org.Name).
				Return(tc.CompanydError)

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

			err = organizationHandler.EnsureDeleted(ctx, org)
			if err != nil {
				t.Fatal(err)
			}

			orgNamespace = &v1.Namespace{}
			err = ctrClient.Get(ctx, client.ObjectKey{Name: namespaceName}, orgNamespace)
			if !k8serrors.IsNotFound(err) {
				t.Fatal(err)
			}
		})
	}

	/*organizationName := "giantswarm"
	//expectedNamespaceName := "org-giantswarm"

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := microloggertest.New()
	k8sClient := unittest.FakeK8sClient()
	ctrClient := k8sClient.CtrlClient()

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
	credentialdClientMock.EXPECT().List(ctx, lister.Request{Organization: organizationName}).Return([]lister.Response{}, nil)

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

	companydClientMock.
		EXPECT().
		DeleteCompany(org.Name).
		Return(nil)

	err = organizationHandler.EnsureDeleted(ctx, org)
	if err != nil {
		t.Fatal(err)
	}*/
}
