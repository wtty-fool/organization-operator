package controllers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/organization-operator/api/v1alpha1"
	"github.com/giantswarm/organization-operator/test/unittest"
)

const (
	testFinalizer = "test.finalizers.giantswarm.io/test-finalizer"
	orgFinalizer  = "operatorkit.giantswarm.io/organization-operator-organization-controller"
)

func Test_OrgReconcileDeleteStep(t *testing.T) {
	testCases := []struct {
		name                string
		organizationName    string
		namespaceFinalizers []string
	}{
		{
			name:                "case 0: Delete org in case there is no org namespace",
			organizationName:    "giantswarm",
			namespaceFinalizers: nil,
		},
		{
			name:                "case 1: Keep org and delete org namespace in case it has no finalizers",
			organizationName:    "giantswarm",
			namespaceFinalizers: []string{},
		},
		{
			name:                "case 2: Keep org and org namespace in case it has finalizers",
			organizationName:    "giantswarm",
			namespaceFinalizers: []string{testFinalizer},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			namespace := newOrgNamespace(tc.organizationName, tc.namespaceFinalizers)
			org := newOrg(tc.organizationName, namespace.Name)

			runtimeObjects := []runtime.Object{org}
			if tc.namespaceFinalizers != nil {
				runtimeObjects = append(runtimeObjects, namespace)
			}

			ctx := context.Background()
			k8sClient := unittest.FakeK8sClient(runtimeObjects...)

			orgController, err := newOrgController(k8sClient)
			if err != nil {
				t.Fatal(err)
			}

			req := newOrgReconcileRequest(org.Name)
			_, err = orgController.Reconcile(ctx, req)
			if err != nil {
				t.Fatal(err)
			}

			org = &v1alpha1.Organization{}
			err = k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Name: tc.organizationName}, org)
			if tc.namespaceFinalizers == nil && (err == nil || !errors.IsNotFound(err)) {
				t.Fatal(err)
			} else if tc.namespaceFinalizers != nil {
				if err != nil {
					t.Fatal(err)
				}
				orgNamespace := &corev1.Namespace{}
				err = k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Name: namespace.Name}, orgNamespace)
				if len(tc.namespaceFinalizers) == 0 && err == nil {
					t.Fatalf("found unexpected namespace %s", namespace.Name)
				} else if len(tc.namespaceFinalizers) == 0 && !errors.IsNotFound(err) {
					t.Fatal(err)
				} else if len(tc.namespaceFinalizers) > 0 && err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func Test_OrgReconcileDelete(t *testing.T) {

	testCases := []struct {
		name                string
		organizationName    string
		namespaceFinalizers []string
		expectedIterations  int
	}{
		{
			name:                "case 0: Delete organization after deleting namespace",
			organizationName:    "giantswarm",
			namespaceFinalizers: []string{testFinalizer},
			expectedIterations:  5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			namespace := newOrgNamespace(tc.organizationName, tc.namespaceFinalizers)
			org := newOrg(tc.organizationName, namespace.Name)

			ctx := context.Background()
			k8sClient := unittest.FakeK8sClient(org, namespace)

			orgController, err := newOrgController(k8sClient)
			if err != nil {
				t.Fatal(err)
			}

			req := newOrgReconcileRequest(org.Name)

			for i := 0; i < tc.expectedIterations; i++ {
				remainingIterations := tc.expectedIterations - i - 1

				orgNamespace := &corev1.Namespace{}
				err = k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Name: namespace.Name}, orgNamespace)

				if remainingIterations == 0 {
					if err == nil {
						t.Fatalf("unexpected namespace %s found", namespace.Name)
					} else if !errors.IsNotFound(err) {
						t.Fatal(err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if remainingIterations == 1 {
						orgNamespace.Finalizers = []string{}
						err = k8sClient.CtrlClient().Update(ctx, orgNamespace)
						if err != nil {
							t.Fatal(err)
						}
					} else {
						if len(orgNamespace.Finalizers) == 0 {
							t.Fatalf("namespace %s has no finalizers", orgNamespace.Name)
						}
					}
					org = &v1alpha1.Organization{}
					err = k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Name: tc.organizationName}, org)
					if err != nil {
						t.Fatal(err)
					}
				}

				_, err = orgController.Reconcile(ctx, req)
				if err != nil {
					t.Fatal(err)
				}
			}

			org = &v1alpha1.Organization{}
			err = k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Name: tc.organizationName}, org)
			if err == nil {
				t.Fatalf("found unexpected organization %s", org.Name)
			} else if !errors.IsNotFound(err) {
				t.Fatal(err)
			}
		})
	}

}

func newOrgController(k8sClient k8sclient.Interface) (*Organization, error) {
	logger := microloggertest.New()

	config := OrganizationConfig{
		K8sClient: k8sClient,
		Logger:    logger,
	}

	return NewOrganization(config)
}
func newOrg(name, namespace string) *v1alpha1.Organization {
	return &v1alpha1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Finalizers:        []string{orgFinalizer},
			DeletionTimestamp: &metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OrganizationSpec{},
		Status: v1alpha1.OrganizationStatus{
			Namespace: namespace,
		},
	}
}

func newOrgNamespace(orgName string, finalizers []string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       fmt.Sprintf("org-%s", orgName),
			Finalizers: finalizers,
		},
	}
}

func newOrgReconcileRequest(orgName string) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: orgName,
		},
	}
}
