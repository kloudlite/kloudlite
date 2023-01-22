package v1

import (
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecondaryEnvSpec defines the desired state of SecondaryEnv
type SecondaryEnvSpec struct {
	PrimaryEnvName string `json:"primaryEnvName"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
// +kubebuilder:resource:scope=Cluster

// SecondaryEnv is the Schema for the secondaryenvs API
type SecondaryEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecondaryEnvSpec `json:"spec,omitempty"`
	Status rApi.Status      `json:"status,omitempty"`
}

func (se *SecondaryEnv) GetStatus() *rApi.Status {
	return &se.Status
}

func (se *SecondaryEnv) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (se *SecondaryEnv) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// SecondaryEnvList contains a list of SecondaryEnv
type SecondaryEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecondaryEnv `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecondaryEnv{}, &SecondaryEnvList{})
}
