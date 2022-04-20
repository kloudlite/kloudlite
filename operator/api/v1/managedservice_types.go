package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManagedServiceSpec defines the desired state of ManagedService
type ManagedServiceSpec struct {
	// Values define the managed services values that correspond to a particular templateId
	Values string `json:"values"`

	// managed svc template Id
	TemplateName string `json:"templateName"`
}

// ManagedServiceStatus defines the observed state of ManagedService
type ManagedServiceStatus struct {
	// Job                  *ReconJob          `json:"job,omitempty"`
	// JobCompleted         *bool              `json:"jobCompleted,omitempty"`
	// DependencyChecked    *map[string]string `json:"dependencyChecked,omitempty"`
	// DeletionJob          *ReconJob          `json:"deletionJob,omitempty"`
	// DeletionJobCompleted *bool              `json:"deletionJobCompleted,omitempty"`
	// NEW
	Generation     int64              `json:"generation,omitempty"`
	ApplyJobCheck  Recon              `json:"apply_job_check,omitempty"`
	DeleteJobCheck Recon              `json:"delete_job_check,omitempty"`
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ManagedService is the Schema for the managedservices API
type ManagedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedServiceSpec   `json:"spec,omitempty"`
	Status ManagedServiceStatus `json:"status,omitempty"`
}

func (msvc *ManagedService) DefaultStatus() {
	msvc.Status.Generation = msvc.Generation
	msvc.Status.ApplyJobCheck = Recon{}
}

func (msvc *ManagedService) IsNewGeneration() bool {
	return msvc.Generation > msvc.Status.Generation
}

func (msvc *ManagedService) HasToBeDeleted() bool {
	return msvc.GetDeletionTimestamp() != nil
}

func (msvc *ManagedService) BuildConditions() {
}

// func (msvc *ManagedService) HasJob() bool {
// 	return msvc.Status.Job != nil && msvc.Status.JobCompleted == nil
// }

// func (msvc *ManagedService) HasNotCheckedDependency() bool {
// 	return false
// 	// return msvc.Status.DependencyChecked == nil
// }

// func (msvc *ManagedService) HasPassedDependencyCheck() bool {
// 	return true
// 	// return msvc.Status.DependencyChecked != nil && len(*msvc.Status.DependencyChecked) == 0
// }

// func (msvc *ManagedService) IsNewGeneration() bool {
// 	return msvc.Status.Generation == nil || msvc.Generation > *msvc.Status.Generation
// }

// func (msvc *ManagedService) ShouldCreateJob() bool {
// 	if msvc.HasPassedDependencyCheck() && msvc.Status.JobCompleted == nil && msvc.Status.Job == nil {
// 		return true
// 	}
// 	return false
// }

// func (msvc *ManagedService) HasToBeDeleted() bool {
// 	return msvc.GetDeletionTimestamp() != nil
// }

// func (msvc *ManagedService) HasDeletionJob() bool {
// 	return msvc.Status.DeletionJob != nil && msvc.Status.DeletionJobCompleted == nil
// }

// func (msvc *ManagedService) ShouldCreateDeletionJob() bool {
// 	if msvc.Status.DeletionJob == nil && msvc.Status.DeletionJobCompleted == nil {
// 		return true
// 	}
// 	return false
// }

//+kubebuilder:object:root=true

// ManagedServiceList contains a list of ManagedService
type ManagedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedService{}, &ManagedServiceList{})
}
