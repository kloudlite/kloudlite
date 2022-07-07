package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rApi "operators.kloudlite.io/lib/operator"
)

type ArtifactRegistry struct {
	Enabled bool `json:"enabled"`
	// specifiy size in GBs
	Size int `json:"size,omitempty"`
}

// ProjectSpec defines the desired state of Project
type ProjectSpec struct {
	// DisplayName of Project
	DisplayName      string           `json:"displayName,omitempty"`
	ArtifactRegistry ArtifactRegistry `json:"artifactRegistry,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// Project is the Schema for the projects API
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (p *Project) LogRef() string {
	return fmt.Sprintf("%s/%s/%s", p.Namespace, p.Kind, p.Name)
}

func (p *Project) GetStatus() *rApi.Status {
	return &p.Status
}

func (p *Project) GetEnsuredLabels() map[string]string {
	return map[string]string{
		fmt.Sprintf("%s/ref", GroupVersion.Group): p.Name,
	}
}

// +kubebuilder:object:root=true

// ProjectList contains a list of Project
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}
