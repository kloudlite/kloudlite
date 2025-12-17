package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PackageSpec defines a Nix package to install
type PackageSpec struct {
	// Name of the package (e.g., nodejs_22, vim, git)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Channel specifies the nixpkgs channel/release to use (e.g., "nixos-24.05", "nixos-23.11", "unstable")
	// Use this for stable, well-known package versions from official releases
	// +optional
	Channel string `json:"channel,omitempty"`

	// NixpkgsCommit specifies an exact nixpkgs commit hash for precise version control
	// Use this when you need a specific historical package version
	// Takes precedence over Channel if both are specified
	// +optional
	NixpkgsCommit string `json:"nixpkgsCommit,omitempty"`
}

// PackageRequestSpec defines the desired packages to install
type PackageRequestSpec struct {
	// WorkspaceRef references the workspace this package request belongs to
	// +kubebuilder:validation:Required
	WorkspaceRef string `json:"workspaceRef"`

	// Packages list of packages to install
	// Empty list is valid (no packages installed yet)
	// +optional
	Packages []PackageSpec `json:"packages,omitempty"`

	// ProfileName is the Nix profile name to use
	// +kubebuilder:validation:Required
	ProfileName string `json:"profileName"`
}

// PackageRequestStatus defines the observed state of PackageRequest
type PackageRequestStatus struct {
	// ObservedGeneration reflects the generation of the most recently observed PackageRequest
	// This allows clients to determine if the status reflects the current spec
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Phase represents the current phase (Pending, Installing, Ready, Failed)
	// +kubebuilder:validation:Enum=Pending;Installing;Ready;Failed
	Phase string `json:"phase,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// ProfileStorePath is the Nix store path of the built environment
	// +optional
	ProfileStorePath string `json:"profileStorePath,omitempty"`

	// PackagesPath is the path to the packages symlink (e.g., /nix/profiles/kloudlite/<workspace>/packages)
	// +optional
	PackagesPath string `json:"packagesPath,omitempty"`

	// SpecHash is a hash of the package specifications for change detection
	// +optional
	SpecHash string `json:"specHash,omitempty"`

	// PackageCount is the number of packages in the environment
	// +optional
	PackageCount int `json:"packageCount,omitempty"`

	// Packages is the list of package names (for display purposes)
	// +optional
	Packages []string `json:"packages,omitempty"`

	// FailedPackage is the name of the package that caused the build to fail (if any)
	// +optional
	FailedPackage string `json:"failedPackage,omitempty"`

	// LastUpdated timestamp of last status update
	// +optional
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={kloudlite,packages}
// +kubebuilder:printcolumn:name="Workspace",type=string,JSONPath=`.spec.workspaceRef`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Packages",type=integer,JSONPath=`.status.packageCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// PackageRequest is the Schema for the packagerequests API
type PackageRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PackageRequestSpec   `json:"spec,omitempty"`
	Status PackageRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PackageRequestList contains a list of PackageRequest
type PackageRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PackageRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PackageRequest{}, &PackageRequestList{})
}
