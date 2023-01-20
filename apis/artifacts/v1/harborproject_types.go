package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/harbor"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

type HarborProjectSpec struct {
	Project *harbor.Project `json:"project,omitempty"`
	Webhook *harbor.Webhook `json:"webhook,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

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
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("HarborProject").String(),
	}
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
