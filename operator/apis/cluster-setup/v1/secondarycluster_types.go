package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rApi "operators.kloudlite.io/pkg/operator"
)


type SecondarySharedConstants struct {
	StatefulPriorityClass string `json:"statefulPriorityClass,omitempty"`
	AppKlAgent string `json:"appKlAgent,omitempty"`
	ImageKlAgent string `json:"imageKlAgent,omitempty"`
}

type SecondaryClusterSpec struct {
	SharedConstants SecondarySharedConstants `json:"SharedConstants,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// SecondaryCluster is the Schema for the secondaryclusters API
type SecondaryCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecondaryClusterSpec `json:"spec,omitempty"`
	Status rApi.Status          `json:"status,omitempty"`
}

func (sc *SecondaryCluster) GetStatus() *rApi.Status {
	return &sc.Status
}

func (sc *SecondaryCluster) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (sc *SecondaryCluster) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// SecondaryClusterList contains a list of SecondaryCluster
type SecondaryClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecondaryCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecondaryCluster{}, &SecondaryClusterList{})
}
