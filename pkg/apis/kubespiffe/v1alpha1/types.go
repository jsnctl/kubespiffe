package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkloadRegistration is a custom resource representing a single kubespiffe
// policy enforcing resource, which selects Pods to issue SVIDs after attestation
type WorkloadRegistration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkloadRegistrationSpec   `json:"spec"`
	Status WorkloadRegistrationStatus `json:"status"`
}

type WorkloadRegistrationSpec struct {
	SPIFFEID string                       `json:"spiffeID"`
	SVIDType string                       `json:"svidType"`
	Selector WorkloadRegistrationSelector `json:"selector,omitempty"`
}

// WorkloadRegistrationSelector constrains which pods may receive an SVID.
// Each field is optional; an empty field matches any value.
type WorkloadRegistrationSelector struct {
	Namespace          string `json:"namespace,omitempty"`
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	PodName            string `json:"podName,omitempty"`
}

type WorkloadRegistrationStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkloadRegistrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []WorkloadRegistration `json:"items"`
}
