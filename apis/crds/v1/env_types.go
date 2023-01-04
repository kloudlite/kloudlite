package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/pkg/constants"
	rApi "operators.kloudlite.io/pkg/operator"
)

type EnvSpec struct {
	ProjectName      string `json:"projectName"`
	BlueprintName    string `json:"blueprintName"`
	AccountRef       string `json:"accountRef"`
	RouterBaseDomain string `json:"routerBaseDomain,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Env is the Schema for the envs API
type Env struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvSpec     `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (e *Env) GetStatus() *rApi.Status {
	return &e.Status
}

func (e *Env) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.ProjectNameKey:   e.Spec.ProjectName,
		constants.BlueprintNameKey: e.Spec.BlueprintName,
		constants.AccountRef:       e.Spec.AccountRef,
	}
}

func (e *Env) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: e.GroupVersionKind().String(),
	}
}

//+kubebuilder:object:root=true

// EnvList contains a list of Env
type EnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Env `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Env{}, &EnvList{})
}
