package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/constants"
	rApi "operators.kloudlite.io/lib/operator"
)

// LambdaSpec defines the desired state of Lambda
type LambdaSpec struct {
	// +kubebuilder:default=1
	// +kubebuilder:validation:Optional
	MinScale int `json:"minScale"`

	// +kubebuilder:default=5
	// +kubebuilder:validation:Optional
	MaxScale int `json:"maxScale"`

	// +kubebuilder:default=100
	// +kubebuilder:validation:Optional
	TargetRps int `json:"targetRps"`

	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

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
		"kloudlite.io/lambda.name": lm.Name,
	}
}

func (m *Lambda) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("Lambda").String(),
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
