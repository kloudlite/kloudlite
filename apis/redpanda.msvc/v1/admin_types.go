package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/pkg/constants"
	rApi "operators.kloudlite.io/pkg/operator"
)

// AdminSpec defines the desired state of Admin
type AdminSpec struct {
	AdminEndpoint string `json:"adminEndpoint"`
	KafkaBrokers  string `json:"kafkaBrokers"`

	Output *ct.Output `json:"output,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Admin is the Schema for the admins API
type Admin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdminSpec   `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (adm *Admin) GetStatus() *rApi.Status {
	return &adm.Status
}

func (adm *Admin) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (adm *Admin) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("Admin").String(),
	}
}

// +kubebuilder:object:root=true

// AdminList contains a list of Admin
type AdminList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Admin `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Admin{}, &AdminList{})
}
