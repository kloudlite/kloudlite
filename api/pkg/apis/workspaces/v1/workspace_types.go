package v1

import (
	corev1 "k8s.io/api/core/v1"
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

// WorkspaceSpec defines the desired state of Workspace
type WorkspaceSpec struct {
	// DisplayName is the human-readable name for the workspace
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=100
	DisplayName string `json:"displayName"`

	// Description provides additional information about the workspace
	// +kubebuilder:validation:MaxLength=500
	// +optional
	Description string `json:"description,omitempty"`

	// Owner is the email/username of the workspace owner
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Owner string `json:"owner"`

	// EnvironmentRef references the default environment for this workspace
	// +optional
	EnvironmentRef *corev1.ObjectReference `json:"environmentRef,omitempty"`

	// MachineTypeRef references the default machine type for this workspace
	// +optional
	MachineTypeRef *corev1.ObjectReference `json:"machineTypeRef,omitempty"`

	// Settings contains workspace-specific settings
	// +optional
	Settings *WorkspaceSettings `json:"settings,omitempty"`

	// Tags for categorizing and filtering workspaces
	// +optional
	Tags []string `json:"tags,omitempty"`

	// ResourceQuota defines resource limits for the workspace
	// +optional
	ResourceQuota *ResourceQuota `json:"resourceQuota,omitempty"`

	// StorageSize is the size of the persistent volume for the workspace
	// +kubebuilder:default="10Gi"
	// +optional
	StorageSize string `json:"storageSize,omitempty"`

	// StorageClassName specifies the storage class to use for the workspace PVC
	// If not specified, uses the cluster's default storage class
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`

	// WorkspacePath is the path inside the workspace container where storage will be mounted
	// +kubebuilder:default=/workspace
	// +optional
	WorkspacePath string `json:"workspacePath,omitempty"`

	// VSCodeVersion specifies the version of VS Code server to use
	// +kubebuilder:default=latest
	// +optional
	VSCodeVersion string `json:"vscodeVersion,omitempty"`

	// ServerType specifies which server to run (code-server, jupyter, ttyd, code-web)
	// +kubebuilder:validation:Enum=code-server;jupyter;ttyd;code-web
	// +kubebuilder:default=code-server
	// +optional
	ServerType string `json:"serverType,omitempty"`

	// Packages list of Nix packages to install in the workspace
	// +optional
	Packages []PackageSpec `json:"packages,omitempty"`

	// Status indicates whether the workspace is active or suspended
	// +kubebuilder:validation:Enum=active;suspended;archived
	// +kubebuilder:default=active
	Status string `json:"status,omitempty"`
}

// WorkspaceSettings contains workspace-specific configuration
type WorkspaceSettings struct {
	// AutoStop indicates if workspace should auto-stop when idle
	// +optional
	AutoStop bool `json:"autoStop,omitempty"`

	// IdleTimeout in minutes before auto-stopping
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10080
	// +optional
	IdleTimeout int32 `json:"idleTimeout,omitempty"`

	// MaxRuntime maximum runtime in minutes before forced stop
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=43200
	// +optional
	MaxRuntime int32 `json:"maxRuntime,omitempty"`

	// StartupScript to run when workspace starts
	// +optional
	StartupScript string `json:"startupScript,omitempty"`

	// EnvironmentVariables to set in the workspace
	// +optional
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`

	// GitConfig for workspace git settings
	// +optional
	GitConfig *GitConfig `json:"gitConfig,omitempty"`

	// VSCodeExtensions list of VS Code extensions to install
	// +optional
	VSCodeExtensions []string `json:"vscodeExtensions,omitempty"`

	// DotfilesRepo URL for dotfiles repository
	// +optional
	DotfilesRepo string `json:"dotfilesRepo,omitempty"`
}

// GitConfig contains git configuration for the workspace
type GitConfig struct {
	// UserName for git commits
	// +optional
	UserName string `json:"userName,omitempty"`

	// UserEmail for git commits
	// +optional
	UserEmail string `json:"userEmail,omitempty"`

	// DefaultBranch name
	// +optional
	DefaultBranch string `json:"defaultBranch,omitempty"`
}

// ResourceQuota defines resource limits for the workspace
type ResourceQuota struct {
	// CPU limit in cores (e.g., "2" or "1000m")
	// +optional
	CPU string `json:"cpu,omitempty"`

	// Memory limit (e.g., "4Gi")
	// +optional
	Memory string `json:"memory,omitempty"`

	// Storage limit (e.g., "100Gi")
	// +optional
	Storage string `json:"storage,omitempty"`

	// GPUs number of GPUs allocated
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=8
	// +optional
	GPUs int32 `json:"gpus,omitempty"`
}

// WorkspaceStatus defines the observed state of Workspace
type WorkspaceStatus struct {
	// Phase represents the current phase of the workspace
	// +kubebuilder:validation:Enum=Pending;Creating;Running;Stopping;Stopped;Failed;Terminating
	Phase string `json:"phase,omitempty"`

	// Conditions represent the latest available observations of workspace state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// LastActivityTime tracks when the workspace was last active
	// +optional
	LastActivityTime *metav1.Time `json:"lastActivityTime,omitempty"`

	// StartTime when the workspace was started
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// StopTime when the workspace was stopped
	// +optional
	StopTime *metav1.Time `json:"stopTime,omitempty"`

	// AccessURL for accessing the workspace
	// +optional
	AccessURL string `json:"accessUrl,omitempty"`

	// ResourceUsage current resource consumption
	// +optional
	ResourceUsage *ResourceUsage `json:"resourceUsage,omitempty"`

	// TotalRuntime in minutes
	// +optional
	TotalRuntime int64 `json:"totalRuntime,omitempty"`

	// PodName is the name of the pod running the VS Code server
	// +optional
	PodName string `json:"podName,omitempty"`

	// PodIP is the IP address of the workspace pod
	// +optional
	PodIP string `json:"podIP,omitempty"`

	// NodeName where the pod is running
	// +optional
	NodeName string `json:"nodeName,omitempty"`

	// InstalledPackages list of successfully installed packages
	// +optional
	InstalledPackages []InstalledPackage `json:"installedPackages,omitempty"`

	// FailedPackages list of packages that failed to install
	// +optional
	FailedPackages []string `json:"failedPackages,omitempty"`

	// PackageInstallationMessage provides information about package installation
	// +optional
	PackageInstallationMessage string `json:"packageInstallationMessage,omitempty"`
}

// ResourceUsage tracks current resource consumption
type ResourceUsage struct {
	// CPU usage in cores
	CPU string `json:"cpu,omitempty"`

	// Memory usage
	Memory string `json:"memory,omitempty"`

	// Storage usage
	Storage string `json:"storage,omitempty"`

	// LastUpdated timestamp
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={kloudlite,workspaces}
// +kubebuilder:printcolumn:name="Display Name",type=string,JSONPath=`.spec.displayName`
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.spec.owner`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.spec.status`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Workspace is the Schema for the workspaces API
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec   `json:"spec,omitempty"`
	Status WorkspaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceList contains a list of Workspace
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}
