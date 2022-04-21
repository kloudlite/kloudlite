package v1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManagedResourceSpec defines the desired state of ManagedResource
type ManagedResourceSpec struct {
	Type       string `json:"type"`
	ManagedSvc string `json:"managedSvc"`
	Values     string `json:"values,omitempty"`
}

// ManagedResourceStatus defines the observed state of ManagedResource
type ManagedResourceStatus struct {
	ApplyJobCheck      Recon              `json:"apply_job_check,omitempty"`
	DeleteJobCheck     Recon              `json:"delete_job_check,omitempty"`
	ManagedSvcDepCheck Recon              `json:"managed_svc_dep_check,omitempty"`
	Generation         int64              `json:"generation,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ManagedResource is the Schema for the managedresources API
type ManagedResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedResourceSpec   `json:"spec,omitempty"`
	Status ManagedResourceStatus `json:"status,omitempty"`
}

func (mres *ManagedResource) DefaultStatus() {
	mres.Status.Generation = mres.Generation
	mres.Status.ApplyJobCheck = Recon{}
}

func (mres *ManagedResource) IsNewGeneration() bool {
	return mres.Generation > mres.Status.Generation
}

func (mres *ManagedResource) HasToBeDeleted() bool {
	return mres.GetDeletionTimestamp() != nil
}

func (mres *ManagedResource) BuildConditions() {
	meta.SetStatusCondition(&mres.Status.Conditions, metav1.Condition{
		Type:               "ApplyJobCheck",
		Status:             mres.Status.ApplyJobCheck.ConditionStatus(),
		ObservedGeneration: mres.Generation,
		Reason:             mres.Status.ApplyJobCheck.Reason(),
		Message:            mres.Status.ApplyJobCheck.Message,
	})

	meta.SetStatusCondition(&mres.Status.Conditions, metav1.Condition{
		Type:               "ManagedSvcDepCheck",
		Status:             mres.Status.ManagedSvcDepCheck.ConditionStatus(),
		ObservedGeneration: mres.Generation,
		Reason:             mres.Status.ManagedSvcDepCheck.Reason(),
		Message:            mres.Status.ManagedSvcDepCheck.Message,
	})

	c := Condition{
		Type:               "Ready",
		Status:             string(metav1.ConditionTrue),
		ObservedGeneration: mres.Generation,
		Reason:             "Success",
		Message:            "all conditions passed",
	}

	for _, cond := range mres.Status.Conditions {
		if cond.Status != metav1.ConditionTrue {
			c.Status = string(cond.Status)
			c.Reason = "ConditionFailed"
			c.Message = fmt.Sprintf("Condition Type=%s Status=%s Message=%s", cond.Type, cond.Status, cond.Message)
			break
		}
	}

	meta.SetStatusCondition(&mres.Status.Conditions, metav1.Condition{
		Type:               c.Type,
		Status:             metav1.ConditionStatus(c.Status),
		ObservedGeneration: mres.Generation,
		Reason:             c.Reason,
		Message:            c.Message,
	})
}

//+kubebuilder:object:root=true

// ManagedResourceList contains a list of ManagedResource
type ManagedResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedResource{}, &ManagedResourceList{})
}
