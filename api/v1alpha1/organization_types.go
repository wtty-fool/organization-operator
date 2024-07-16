/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OrganizationSpec defines the desired state of Organization
type OrganizationSpec struct {
	// Add any additional fields if needed
}

// OrganizationStatus defines the observed state of Organization
type OrganizationStatus struct {
	// Namespace is the namespace containing the resources for this organization.
	Namespace string `json:"namespace,omitempty"`
}

//nolint:revive
//+kubebuilder:object:root=true
//nolint:revive
//+kubebuilder:subresource:status
//nolint:revive
//+kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".status.namespace"
//nolint:revive
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//nolint:revive
//+kubebuilder:resource:scope=Cluster,categories={common,giantswarm},shortName={org,orgs}

// Organization represents schema for managed Kubernetes namespace.
// Reconciled by organization-operator.
type Organization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrganizationSpec   `json:"spec,omitempty"`
	Status OrganizationStatus `json:"status,omitempty"`
}

//nolint:revive
//+kubebuilder:object:root=true

// OrganizationList contains a list of Organization
type OrganizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Organization `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Organization{}, &OrganizationList{})
}
