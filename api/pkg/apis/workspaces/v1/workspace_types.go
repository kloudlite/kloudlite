package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	// WorkMachineRef references the WorkMachine where this workspace runs
	// +kubebuilder:validation:Required
	WorkMachineRef *corev1.ObjectReference `json:"workMachineRef"`

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

	// AssignedWorkMachine references the actual WorkMachine resource that was assigned
	// +optional
	AssignedWorkMachine *corev1.ObjectReference `json:"assignedWorkMachine,omitempty"`
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
