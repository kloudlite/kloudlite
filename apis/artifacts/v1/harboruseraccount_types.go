package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HarborUserAccountSpec defines the desired state of HarborUserAccount
type HarborUserAccountSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of HarborUserAccount. Edit harboruseraccount_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// HarborUserAccountStatus defines the observed state of HarborUserAccount
type HarborUserAccountStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HarborUserAccount is the Schema for the harboruseraccounts API
type HarborUserAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborUserAccountSpec   `json:"spec,omitempty"`
	Status HarborUserAccountStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HarborUserAccountList contains a list of HarborUserAccount
type HarborUserAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborUserAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HarborUserAccount{}, &HarborUserAccountList{})
}
