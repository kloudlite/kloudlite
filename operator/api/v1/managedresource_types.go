package v1

import (
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
	Job                  *ReconJob          `json:"job,omitempty"`
	JobCompleted         *bool              `json:"jobCompleted,omitempty"`
	Generation           *int64             `json:"generation,omitempty"`
	DependencyChecked    *map[string]string `json:"dependencyChecked,omitempty"`
	DeletionJob          *ReconJob          `json:"deletionJob,omitempty"`
	DeletionJobCompleted *bool              `json:"deletionJobCompleted,omitempty"`
	Conditions           []metav1.Condition `json:"conditions,omitempty"`
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
	mres.Status.DependencyChecked = nil
	mres.Status.Job = nil
	mres.Status.JobCompleted = nil
	mres.Status.Generation = &mres.Generation
}

func (mres *ManagedResource) HasJob() bool {
	return mres.Status.Job != nil && mres.Status.JobCompleted == nil
}

func (mres *ManagedResource) HasNotCheckedDependency() bool {
	return mres.Status.DependencyChecked == nil
}

func (mres *ManagedResource) HasPassedDependencyCheck() bool {
	return mres.Status.DependencyChecked != nil && len(*mres.Status.DependencyChecked) == 0
}

func (mres *ManagedResource) IsNewGeneration() bool {
	return mres.Status.Generation == nil || mres.Generation > *mres.Status.Generation
}

func (mres *ManagedResource) ShouldCreateJob() bool {
	if mres.HasPassedDependencyCheck() && mres.Status.JobCompleted == nil && mres.Status.Job == nil {
		return true
	}
	return false
}

func (mres *ManagedResource) HasToBeDeleted() bool {
	return mres.GetDeletionTimestamp() != nil
}

func (mres *ManagedResource) HasDeletionJob() bool {
	return mres.Status.DeletionJob != nil && mres.Status.DeletionJobCompleted == nil
}

func (mres *ManagedResource) ShouldCreateDeletionJob() bool {
	if mres.Status.DeletionJob == nil && mres.Status.DeletionJobCompleted == nil {
		return true
	}
	return false
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
