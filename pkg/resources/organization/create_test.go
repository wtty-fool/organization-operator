package organization

import (
	"context"
	"testing"

	"github.com/go-logr/logr/testr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
)

func TestNamespaceIsCreated(t *testing.T) {
	// Set up a new scheme that includes our custom APIs
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = securityv1alpha1.AddToScheme(scheme)

	// Test parameters
	organizationName := "giantswarm"
	expectedNamespaceName := "org-giantswarm"

	// Create a new organization
	org := &securityv1alpha1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: organizationName,
		},
		Spec: securityv1alpha1.OrganizationSpec{},
	}

	// Create a fake client
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(org).Build()

	// Create the Resource
	r := &Resource{
		k8sClient: fakeClient,
		logger:    testr.New(t),
	}

	// Run EnsureCreated
	ctx := context.Background()
	err := r.EnsureCreated(ctx, org)
	if err != nil {
		t.Fatalf("EnsureCreated failed: %v", err)
	}

	// Check if the namespace was created
	orgNamespace := &corev1.Namespace{}
	err = fakeClient.Get(ctx, client.ObjectKey{Name: expectedNamespaceName}, orgNamespace)
	if err != nil {
		t.Fatalf("Failed to get created namespace: %v", err)
	}

	// Check if the organization status was updated
	updatedOrg := &securityv1alpha1.Organization{}
	err = fakeClient.Get(ctx, client.ObjectKey{Name: organizationName}, updatedOrg)
	if err != nil {
		t.Fatalf("Failed to get updated organization: %v", err)
	}

	if updatedOrg.Status.Namespace != expectedNamespaceName {
		t.Fatalf("Created namespace should be stored in organization status. Expected %s, got %s",
			expectedNamespaceName, updatedOrg.Status.Namespace)
	}

	// Check if the namespace has the correct labels
	if orgNamespace.Labels["giantswarm.io/organization"] != organizationName {
		t.Fatalf("Namespace should have correct organization label. Expected %s, got %s",
			organizationName, orgNamespace.Labels["giantswarm.io/organization"])
	}

	if orgNamespace.Labels["giantswarm.io/managed-by"] != "organization-operator" {
		t.Fatalf("Namespace should have correct managed-by label. Expected organization-operator, got %s",
			orgNamespace.Labels["giantswarm.io/managed-by"])
	}
}
