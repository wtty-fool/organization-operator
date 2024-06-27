package controllers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
)

const (
	testFinalizer = "test.finalizers.giantswarm.io/test-finalizer"
	orgFinalizer  = "operatorkit.giantswarm.io/organization-operator-organization-controller"
)

func TestOrganizationReconciler_Reconcile(t *testing.T) {
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

			scheme := runtime.NewScheme()
			_ = clientgoscheme.AddToScheme(scheme)
			_ = securityv1alpha1.AddToScheme(scheme)

			objs := []runtime.Object{org}
			if tc.namespaceFinalizers != nil {
				objs = append(objs, namespace)
			}

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build()
			reconciler := &OrganizationReconciler{
				Client: fakeClient,
				Log:    testr.New(t),
				Scheme: scheme,
			}

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: tc.organizationName,
				},
			}

			_, err := reconciler.Reconcile(context.Background(), req)
			if err != nil {
				t.Fatalf("Reconcile failed: %v", err)
			}

			var resultOrg securityv1alpha1.Organization
			err = fakeClient.Get(context.Background(), client.ObjectKey{Name: tc.organizationName}, &resultOrg)
			if tc.namespaceFinalizers == nil && !errors.IsNotFound(err) {
				t.Fatalf("Expected organization to be deleted, but got error: %v", err)
			} else if tc.namespaceFinalizers != nil {
				if err != nil {
					t.Fatalf("Failed to get organization: %v", err)
				}

				var resultNamespace corev1.Namespace
				err = fakeClient.Get(context.Background(), client.ObjectKey{Name: namespace.Name}, &resultNamespace)
				if len(tc.namespaceFinalizers) == 0 && !errors.IsNotFound(err) {
					t.Fatalf("Expected namespace to be deleted, but got error: %v", err)
				} else if len(tc.namespaceFinalizers) > 0 && err != nil {
					t.Fatalf("Failed to get namespace: %v", err)
				}
			}
		})
	}
}

func TestOrganizationReconciler_ReconcileDelete(t *testing.T) {
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

			scheme := runtime.NewScheme()
			_ = clientgoscheme.AddToScheme(scheme)
			_ = securityv1alpha1.AddToScheme(scheme)

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(org, namespace).Build()
			reconciler := &OrganizationReconciler{
				Client: fakeClient,
				Log:    testr.New(t),
				Scheme: scheme,
			}

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: tc.organizationName,
				},
			}

			for i := 0; i < tc.expectedIterations; i++ {
				remainingIterations := tc.expectedIterations - i - 1

				var resultNamespace corev1.Namespace
				err := fakeClient.Get(context.Background(), client.ObjectKey{Name: namespace.Name}, &resultNamespace)

				if remainingIterations == 0 {
					if !errors.IsNotFound(err) {
						t.Fatalf("Expected namespace to be deleted, but got error: %v", err)
					}
				} else {
					if err != nil {
						t.Fatalf("Failed to get namespace: %v", err)
					}
					if remainingIterations == 1 {
						resultNamespace.Finalizers = []string{}
						err = fakeClient.Update(context.Background(), &resultNamespace)
						if err != nil {
							t.Fatalf("Failed to update namespace: %v", err)
						}
					} else {
						if len(resultNamespace.Finalizers) == 0 {
							t.Fatalf("Namespace %s has no finalizers", resultNamespace.Name)
						}
					}

					var resultOrg securityv1alpha1.Organization
					err = fakeClient.Get(context.Background(), client.ObjectKey{Name: tc.organizationName}, &resultOrg)
					if err != nil {
						t.Fatalf("Failed to get organization: %v", err)
					}
				}

				_, err = reconciler.Reconcile(context.Background(), req)
				if err != nil {
					t.Fatalf("Reconcile failed: %v", err)
				}
			}

			var resultOrg securityv1alpha1.Organization
			err := fakeClient.Get(context.Background(), client.ObjectKey{Name: tc.organizationName}, &resultOrg)
			if !errors.IsNotFound(err) {
				t.Fatalf("Expected organization to be deleted, but got error: %v", err)
			}
		})
	}
}

func newOrg(name, namespace string) *securityv1alpha1.Organization {
	return &securityv1alpha1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Finalizers:        []string{orgFinalizer},
			DeletionTimestamp: &metav1.Time{Time: time.Now()},
		},
		Spec: securityv1alpha1.OrganizationSpec{},
		Status: securityv1alpha1.OrganizationStatus{
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
