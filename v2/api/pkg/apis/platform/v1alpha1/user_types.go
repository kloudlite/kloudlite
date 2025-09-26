package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProviderAccount represents an OAuth provider account
type ProviderAccount struct {
	// Provider name (google, github, microsoft-entra-id)
	Provider string `json:"provider"`

	// Provider-specific user ID
	ProviderID string `json:"providerId"`

	// Email from provider
	Email string `json:"email"`

	// Name from provider
	// +optional
	Name string `json:"name,omitempty"`

	// Avatar URL from provider
	// +optional
	Image string `json:"image,omitempty"`

	// When this provider was connected
	ConnectedAt metav1.Time `json:"connectedAt"`
}

// UserSpec defines the desired state of User
type UserSpec struct {
	// Email address of the user (primary identifier)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	Email string `json:"email"`

	// Display name of the user
	// +kubebuilder:validation:MaxLength=100
	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// URL to user's avatar image
	// +optional
	AvatarURL string `json:"avatarUrl,omitempty"`

	// OAuth provider accounts linked to this user
	// +optional
	Providers []ProviderAccount `json:"providers,omitempty"`

	// Roles of the user in the platform
	// +optional
	Roles []string `json:"roles,omitempty"`

	// Whether the user account is active
	// +kubebuilder:default=true
	// +optional
	Active *bool `json:"active,omitempty"`

	// Additional metadata
	// +optional
	Metadata map[string]string `json:"metadata,omitempty"`
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
// +kubebuilder:printcolumn:name="Email",type="string",JSONPath=".spec.email"
// +kubebuilder:printcolumn:name="DisplayName",type="string",JSONPath=".spec.displayName"
// +kubebuilder:printcolumn:name="Providers",type="string",JSONPath=".spec.providers[*].provider"
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