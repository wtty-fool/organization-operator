package organization

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/organization-operator/api/v1alpha1"
	"github.com/giantswarm/organization-operator/test/unittest"
)

func Test_NamespaceIsDeleted(t *testing.T) {

	testCases := []struct {
		Name             string
		OrganizationName string
	}{
		{
			Name:             "case 0: Delete org with valid response",
			OrganizationName: "giantswarm",
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
			config := Config{
				K8sClient: k8sClient,
				Logger:    logger,
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
}
