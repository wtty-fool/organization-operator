package organization

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr/testr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
)

func TestNamespaceIsDeleted(t *testing.T) {
	testCases := []struct {
		name             string
		organizationName string
	}{
		{
			name:             "case 0: Delete org with valid response",
			organizationName: "giantswarm",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up a new scheme that includes our custom APIs
			scheme := runtime.NewScheme()
			_ = clientgoscheme.AddToScheme(scheme)
			_ = securityv1alpha1.AddToScheme(scheme)

			namespaceName := fmt.Sprintf("org-%s", tc.organizationName)

			org := &securityv1alpha1.Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name: tc.organizationName,
				},
				Spec: securityv1alpha1.OrganizationSpec{},
				Status: securityv1alpha1.OrganizationStatus{
					Namespace: namespaceName,
				},
			}

			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
				},
			}

			// Create a fake client with the organization and namespace
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(org, namespace).
				Build()

			// Create the Resource
			r := &Resource{
				k8sClient: fakeClient,
				logger:    testr.New(t),
			}

			// Run EnsureDeleted
			ctx := context.Background()
			err := r.EnsureDeleted(ctx, org)
			if err != nil {
				t.Fatalf("EnsureDeleted failed: %v", err)
			}

			// Check if the namespace was deleted
			deletedNamespace := &corev1.Namespace{}
			err = fakeClient.Get(ctx, client.ObjectKey{Name: namespaceName}, deletedNamespace)
			if !apierrors.IsNotFound(err) {
				t.Fatalf("Expected namespace to be deleted, but got error: %v", err)
			}

			// Optionally, check if the organization's status was updated
			updatedOrg := &securityv1alpha1.Organization{}
			err = fakeClient.Get(ctx, client.ObjectKey{Name: tc.organizationName}, updatedOrg)
			if err != nil {
				t.Fatalf("Failed to get updated organization: %v", err)
			}

			// Add any additional checks for the organization's status here
			// For example, you might want to check if the Namespace field in the status is cleared
			if updatedOrg.Status.Namespace != "" {
				t.Fatalf("Expected organization status namespace to be empty, but got: %s", updatedOrg.Status.Namespace)
			}
		})
	}
}
