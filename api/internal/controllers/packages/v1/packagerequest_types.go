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

// InstalledPackage represents a successfully installed package
type InstalledPackage struct {
	// Name of the package
	Name string `json:"name"`

	// Version of the installed package
	// +optional
	Version string `json:"version,omitempty"`

	// BinPath where binaries are located
	// +optional
	BinPath string `json:"binPath,omitempty"`

	// StorePath in the Nix store
	// +optional
	StorePath string `json:"storePath,omitempty"`

	// InstalledAt timestamp
	// +optional
	InstalledAt metav1.Time `json:"installedAt,omitempty"`
}

// PackageRequestSpec defines the desired packages to install
type PackageRequestSpec struct {
	// WorkspaceRef references the workspace this package request belongs to
	// +kubebuilder:validation:Required
	WorkspaceRef string `json:"workspaceRef"`

	// Packages list of packages to install
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Packages []PackageSpec `json:"packages"`

	// ProfileName is the Nix profile name to use
	// +kubebuilder:validation:Required
	ProfileName string `json:"profileName"`
}

// PackageRequestStatus defines the observed state of PackageRequest
type PackageRequestStatus struct {
	// Phase represents the current phase (Pending, Installing, Ready, Failed)
	// +kubebuilder:validation:Enum=Pending;Installing;Ready;Failed
	Phase string `json:"phase,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// InstalledPackages list of successfully installed packages
	// +optional
	InstalledPackages []InstalledPackage `json:"installedPackages,omitempty"`

	// FailedPackages list of packages that failed to install
	// NOTE: No omitempty - empty slice must be serialized to clear old failures
	// +optional
	FailedPackages []string `json:"failedPackages"`

	// LastUpdated timestamp of last status update
	// +optional
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={kloudlite,packages}
// +kubebuilder:printcolumn:name="Workspace",type=string,JSONPath=`.spec.workspaceRef`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Installed",type=integer,JSONPath=`.status.installedPackages[*].name`
// +kubebuilder:printcolumn:name="Failed",type=integer,JSONPath=`.status.failedPackages[*]`
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
