package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "operators.kloudlite.io/apis/crds/v1"
	rApi "operators.kloudlite.io/lib/operator"
)

// LambdaSpec defines the desired state of Lambda
type LambdaSpec struct {
	Containers []v1.AppContainer `json:"containers,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Lambda is the Schema for the lambdas API
type Lambda struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LambdaSpec  `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (lm *Lambda) GetStatus() *rApi.Status {
	return &lm.Status
}

func (lm *Lambda) GetEnsuredLabels() map[string]string {
	return map[string]string{
		fmt.Sprintf("%s/ref", GroupVersion.Group): lm.Name,
	}
}

// +kubebuilder:object:root=true

// LambdaList contains a list of Lambda
type LambdaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Lambda `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Lambda{}, &LambdaList{})
}
