package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceReference represents a reference to a pinned resource
type ResourceReference struct {
	// Name of the resource
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the resource (for namespaced resources like Workspaces)
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// UserPreferencesSpec defines the desired state of UserPreferences
type UserPreferencesSpec struct {
	// PinnedWorkspaces is a list of pinned workspace references
	// +optional
	PinnedWorkspaces []ResourceReference `json:"pinnedWorkspaces,omitempty"`

	// PinnedEnvironments is a list of pinned environment names
	// +optional
	PinnedEnvironments []string `json:"pinnedEnvironments,omitempty"`
}

// UserPreferencesStatus defines the observed state of UserPreferences
type UserPreferencesStatus struct {
	// LastUpdated is when preferences were last modified
	// +optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=uprefs
// +kubebuilder:printcolumn:name="Pinned Workspaces",type="integer",JSONPath=".spec.pinnedWorkspaces",description="Number of pinned workspaces"
// +kubebuilder:printcolumn:name="Pinned Environments",type="integer",JSONPath=".spec.pinnedEnvironments",description="Number of pinned environments"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// UserPreferences is the Schema for the userpreferences API
// The resource name should match the username (User.metadata.name)
type UserPreferences struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserPreferencesSpec   `json:"spec,omitempty"`
	Status UserPreferencesStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserPreferencesList contains a list of UserPreferences
type UserPreferencesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserPreferences `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UserPreferences{}, &UserPreferencesList{})
}
