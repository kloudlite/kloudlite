package v1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProjectSpec defines the desired state of Project
type ProjectSpec struct {
	// DisplayName of Project
	DisplayName string `json:"displayName,omitempty"`
}

// ProjectStatus defines the observed state of Project
type ProjectStatus struct {
	NamespaceCheck    Recon              `json:"namespace,omitempty"`
	DelNamespaceCheck Recon              `json:"del_namespace_check,omitempty"`
	Generation        int64              `json:"generation,omitempty"`
	Conditions        []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Project is the Schema for the projects API
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec   `json:"spec,omitempty"`
	Status ProjectStatus `json:"status,omitempty"`
}

func (p *Project) DefaultStatus() {
	p.Status.Generation = p.Generation
	p.Status.NamespaceCheck = Recon{}
}

func (p *Project) IsNewGeneration() bool {
	return p.Generation > p.Status.Generation
}

func (p *Project) HasToBeDeleted() bool {
	return p.GetDeletionTimestamp() != nil
}

func (p *Project) BuildConditions() {
	meta.SetStatusCondition(&p.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             p.Status.NamespaceCheck.ConditionStatus(),
		ObservedGeneration: p.Generation,
		Reason:             p.Status.NamespaceCheck.Reason(),
		Message:            p.Status.NamespaceCheck.Message,
	})
}

//+kubebuilder:object:root=true

// ProjectList contains a list of Project
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}
