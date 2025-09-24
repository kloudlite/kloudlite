package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UserSpec defines the desired state of User
type UserSpec struct {
	// Email address of the user
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	Email string `json:"email"`

	// Unique username for the user
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=30
	// +kubebuilder:validation:Pattern=`^[a-z0-9][a-z0-9-]*[a-z0-9]$`
	Username string `json:"username"`

	// Display name of the user
	// +kubebuilder:validation:MaxLength=100
	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// URL to user's avatar image
	// +optional
	AvatarURL string `json:"avatarUrl,omitempty"`

	// Role of the user in the platform
	// +kubebuilder:validation:Enum=admin;developer;viewer
	// +kubebuilder:default=developer
	// +optional
	Role string `json:"role,omitempty"`

	// Whether the user account is active
	// +kubebuilder:default=true
	// +optional
	Active *bool `json:"active,omitempty"`
}

// UserStatus defines the observed state of User
type UserStatus struct {
	// Current status of the user
	// +kubebuilder:validation:Enum=active;inactive;suspended;pending
	// +optional
	Phase string `json:"phase,omitempty"`

	// Last login timestamp
	// +optional
	LastLogin *metav1.Time `json:"lastLogin,omitempty"`

	// User creation timestamp
	// +optional
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=usr
// +kubebuilder:printcolumn:name="Username",type="string",JSONPath=".spec.username"
// +kubebuilder:printcolumn:name="Email",type="string",JSONPath=".spec.email"
// +kubebuilder:printcolumn:name="Role",type="string",JSONPath=".spec.role"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// User is the Schema for the users API
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}