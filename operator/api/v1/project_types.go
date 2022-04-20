package v1

import (
	"fmt"

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
	SvcAccountCheck   Recon              `json:"svc_account_check,omitempty"`
	PullSecretCheck   Recon              `json:"pull_secret_check,omitempty"`
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
	p.Status.SvcAccountCheck = Recon{}
	p.Status.PullSecretCheck = Recon{}
}

func (p *Project) IsNewGeneration() bool {
	return p.Generation > p.Status.Generation
}

func (p *Project) HasToBeDeleted() bool {
	return p.GetDeletionTimestamp() != nil
}

func (p *Project) BuildConditions() {
	// NamespaceCheck
	meta.SetStatusCondition(&p.Status.Conditions, metav1.Condition{
		Type:               "NamespaceCheck",
		Status:             p.Status.NamespaceCheck.ConditionStatus(),
		ObservedGeneration: p.Generation,
		Reason:             p.Status.NamespaceCheck.Reason(),
		Message:            p.Status.NamespaceCheck.Message,
	})

	meta.SetStatusCondition(&p.Status.Conditions, metav1.Condition{
		Type:               "PullSecretCheck",
		Status:             p.Status.PullSecretCheck.ConditionStatus(),
		ObservedGeneration: p.Generation,
		Reason:             p.Status.PullSecretCheck.Reason(),
		Message:            p.Status.PullSecretCheck.Message,
	})

	// SvcAccountCheck
	meta.SetStatusCondition(&p.Status.Conditions, metav1.Condition{
		Type:               "SvcAccountCheck",
		Status:             p.Status.SvcAccountCheck.ConditionStatus(),
		ObservedGeneration: p.Generation,
		Reason:             p.Status.SvcAccountCheck.Reason(),
		Message:            p.Status.SvcAccountCheck.Message,
	})

	x := Condition{
		Status:  string(metav1.ConditionTrue),
		Reason:  "Success",
		Message: "all conditions checks passed",
	}
	for _, c := range p.Status.Conditions {
		if c.Type != "Ready" && c.Status != metav1.ConditionTrue {
			x.Status = string(metav1.ConditionFalse)
			x.Reason = "Failed"
			x.Message = fmt.Sprintf("Type=%s Status=%s", c.Type, c.Status)
		}
	}

	// Ready
	meta.SetStatusCondition(&p.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionStatus(x.Status),
		ObservedGeneration: p.Generation,
		Reason:             x.Reason,
		Message:            x.Message,
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
