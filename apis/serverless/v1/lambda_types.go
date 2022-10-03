package v1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
)

// LambdaSpec defines the desired state of Lambda
type LambdaSpec struct {
	// +kubebuilder:default=kloudlite-svc-account
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// +kubebuilder:default=1
	MinScale int `json:"minScale,omitempty"`

	// +kubebuilder:default=5
	MaxScale int `json:"maxScale,omitempty"`

	// +kubebuilder:default=100
	TargetRps int `json:"targetRps,omitempty"`

	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

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
	m := map[string]string{
		constants.LambdaNameKey: lm.Name,
	}

	for idx := range lm.Spec.Containers {
		m[fmt.Sprintf("kloudlite.io/image-%s", fn.Sha1Sum([]byte(lm.Spec.Containers[idx].Image)))] = "true"
	}
	return m
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
