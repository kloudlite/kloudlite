package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

type AnchorSpec struct {
	Type      string                  `json:"type"`
	ParentGVK metav1.GroupVersionKind `json:"parentGVK"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Anchor is the Schema for the anchors API
type Anchor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AnchorSpec  `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AnchorList contains a list of Anchor
type AnchorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Anchor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Anchor{}, &AnchorList{})
}
