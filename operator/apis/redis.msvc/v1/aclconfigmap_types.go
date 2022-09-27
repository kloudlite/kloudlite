package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/lib/constants"
	rApi "operators.kloudlite.io/lib/operator"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ACLConfigMapSpec defines the desired state of ACLConfigMap
type ACLConfigMapSpec struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status-watcher

// ACLConfigMap is the Schema for the aclconfigmaps API
type ACLConfigMap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ACLConfigMapSpec `json:"spec,omitempty"`
	Status rApi.Status      `json:"status-watcher,omitempty"`
}

func (cfg *ACLConfigMap) GetStatus() *rApi.Status {
	return &cfg.Status
}

func (cfg *ACLConfigMap) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.MsvcNameKey: cfg.Name,
	}
}

func (cfg *ACLConfigMap) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// ACLConfigMapList contains a list of ACLConfigMap
type ACLConfigMapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ACLConfigMap `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ACLConfigMap{}, &ACLConfigMapList{})
}
