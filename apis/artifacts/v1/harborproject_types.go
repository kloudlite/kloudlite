package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rApi "operators.kloudlite.io/lib/operator"
)

type HarborProjectSpec struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// HarborProject is the Schema for the harborprojects API
type HarborProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborProjectSpec `json:"spec,omitempty"`
	Status rApi.Status       `json:"status,omitempty"`
}

func (hp *HarborProject) GetStatus() *rApi.Status {
	return &hp.Status
}

func (hp *HarborProject) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (in *HarborProject) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// HarborProjectList contains a list of HarborProject
type HarborProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborProject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HarborProject{}, &HarborProjectList{})
}
