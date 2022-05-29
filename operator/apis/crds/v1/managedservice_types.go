package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	fn "operators.kloudlite.io/lib/functions"
	libOperator "operators.kloudlite.io/lib/operator"
	rawJson "operators.kloudlite.io/lib/raw-json"
)

// ManagedServiceSpec defines the desired state of ManagedService
type ManagedServiceSpec struct {
	ApiVersion string              `json:"apiVersion"`
	Inputs     rawJson.KubeRawJson `json:"inputs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ManagedService is the Schema for the managedservices API
type ManagedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedServiceSpec `json:"spec,omitempty"`
	Status libOperator.Status `json:"status,omitempty"`
}

func (m *ManagedService) NameRef() string {
	return fmt.Sprintf("%s/%s/%s", m.Namespace, m.Kind, m.Name)
}

func (m *ManagedService) GetStatus() *libOperator.Status {
	return &m.Status
}

func (m *ManagedService) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (m *ManagedService) GetWatchLabels() map[string]string {
	return map[string]string{
		"msvc.kloudlite.io/ref": m.Name,
	}
}

func (s *ManagedService) Hash() (string, error) {
	m := make(map[string]interface{}, 3)
	m["name"] = s.Name
	m["namespace"] = s.Namespace
	m["spec"] = s.Spec
	hash, err := fn.Json.Hash(m)
	return hash, err
}

// +kubebuilder:object:root=true

// ManagedServiceList contains a list of ManagedService
type ManagedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedService{}, &ManagedServiceList{})
}
