package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConnectionTokenSpec defines the desired state of ConnectionToken
type ConnectionTokenSpec struct {
	// Display name for the token
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=100
	DisplayName string `json:"displayName"`

	// User ID (email) who owns this token
	// +kubebuilder:validation:Required
	UserID string `json:"userId"`

	// SSH jump host for connecting to workspaces
	// +kubebuilder:validation:Required
	SSHJumpHost string `json:"sshJumpHost"`

	// SSH port for jump host
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	SSHPort int `json:"sshPort"`

	// API URL for accessing Kloudlite API
	// +kubebuilder:validation:Required
	APIURL string `json:"apiUrl"`

	// Token expiration time (optional, tokens don't expire by default)
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`
}

// ConnectionTokenStatus defines the observed state of ConnectionToken
type ConnectionTokenStatus struct {
	// Whether the token is ready to use
	// +optional
	IsReady bool `json:"isReady,omitempty"`

	// Status message
	// +optional
	Message string `json:"message,omitempty"`

	// Last time the token was used
	// +optional
	LastUsed *metav1.Time `json:"lastUsed,omitempty"`

	// JWT token (only populated on creation, cleared after first read)
	// This field is ephemeral and will be cleared after the user retrieves it
	// +optional
	Token string `json:"token,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=ct,singular=connectiontoken
// +kubebuilder:printcolumn:name="DisplayName",type="string",JSONPath=".spec.displayName"
// +kubebuilder:printcolumn:name="UserID",type="string",JSONPath=".spec.userId"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.isReady"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ConnectionToken is the Schema for the connectiontokens API
type ConnectionToken struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectionTokenSpec   `json:"spec,omitempty"`
	Status ConnectionTokenStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConnectionTokenList contains a list of ConnectionToken
type ConnectionTokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConnectionToken `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConnectionToken{}, &ConnectionTokenList{})
}
