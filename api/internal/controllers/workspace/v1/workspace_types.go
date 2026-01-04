package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnvironmentConnectionSpec defines the environment connection
type EnvironmentConnectionSpec struct {
	// EnvironmentRef references the environment to connect to
	// +kubebuilder:validation:Required
	EnvironmentRef corev1.ObjectReference `json:"environmentRef"`
}

// ExposedPort defines a port to expose from the workspace
// All exposed ports get an HTTP ingress route with hostname p{port}-{hash}.{subdomain}
type ExposedPort struct {
	// Port is the port number to expose
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
}

// GitRepository defines a git repository to clone when workspace starts
type GitRepository struct {
	// URL of the git repository (supports https:// and git@ formats)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	URL string `json:"url"`

	// Branch to clone (optional, uses repository default if not specified)
	// +optional
	Branch string `json:"branch,omitempty"`
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

	// OwnedBy is the username of the workspace owner
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	OwnedBy string `json:"ownedBy"`

	// Visibility controls who can see this workspace
	// - private: only the owner can see
	// - shared: shared with specific users listed in SharedWith
	// - open: visible to all team members
	// +kubebuilder:validation:Enum=private;shared;open
	// +kubebuilder:default=private
	// +optional
	Visibility string `json:"visibility,omitempty"`

	// SharedWith is the list of usernames this workspace is shared with
	// Only used when Visibility is "shared"
	// +optional
	SharedWith []string `json:"sharedWith,omitempty"`

	// WorkmachineName references the WorkMachine this workspace belongs to
	// The workspace will run in the WorkMachine's targetNamespace
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	WorkmachineName string `json:"workmachine"`

	// EnvironmentConnection defines the environment to connect to and associated intercepts
	// When set to nil, workspace is disconnected and all intercepts are removed
	// +optional
	EnvironmentConnection *EnvironmentConnectionSpec `json:"environmentConnection,omitempty"`

	// GitRepository defines a git repository to clone when workspace starts
	// The repository will be cloned into the workspace folder using SSH keys from the WorkMachine
	// +optional
	GitRepository *GitRepository `json:"gitRepository,omitempty"`

	// Settings contains workspace-specific settings
	// +optional
	Settings *WorkspaceSettings `json:"settings,omitempty"`

	// Tags for categorizing and filtering workspaces
	// +optional
	Tags []string `json:"tags,omitempty"`

	// ResourceQuota defines resource limits for the workspace
	// +optional
	ResourceQuota *ResourceQuota `json:"resourceQuota,omitempty"`

	// VSCodeVersion specifies the version of VS Code server to use
	// +kubebuilder:default=latest
	// +optional
	VSCodeVersion string `json:"vscodeVersion,omitempty"`

	// Status indicates whether the workspace is active or suspended
	// +kubebuilder:validation:Enum=active;suspended;archived
	// +kubebuilder:default=active
	Status string `json:"status,omitempty"`

	// FromSnapshot specifies a pushed snapshot to create this workspace from
	// Only snapshots with status.registryStatus.pushed=true can be used
	// This field is automatically cleared after successful restoration
	// +optional
	FromSnapshot *FromSnapshotRef `json:"fromSnapshot,omitempty"`

	// Expose defines ports to expose from the workspace
	// Each port gets an HTTP ingress route with hostname p{port}-{hash}.{subdomain}
	// +optional
	Expose []ExposedPort `json:"expose,omitempty"`
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

// ConnectedEnvironmentInfo tracks the connected environment details
// If this struct exists in status, it means the workspace is connected to the environment
type ConnectedEnvironmentInfo struct {
	// Name of the connected environment
	Name string `json:"name"`

	// TargetNamespace where environment services are deployed
	TargetNamespace string `json:"targetNamespace"`

	// AvailableServices lists services available in the environment
	// +optional
	AvailableServices []string `json:"availableServices,omitempty"`
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

	// AccessURL for accessing the workspace (deprecated, use AccessURLs instead)
	// +optional
	AccessURL string `json:"accessUrl,omitempty"`

	// AccessURLs contains URLs for accessing different workspace services
	// Keys are service names (code-server, ttyd, jupyter, code-web)
	// +optional
	AccessURLs map[string]string `json:"accessUrls,omitempty"`

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

	// ActiveConnections is the number of active network connections in the workspace
	// +optional
	ActiveConnections int `json:"activeConnections,omitempty"`

	// IdleState tracks whether the workspace is idle or active
	// +kubebuilder:validation:Enum=active;idle
	// +optional
	IdleState string `json:"idleState,omitempty"`

	// IdleSince tracks when the workspace became idle
	// +optional
	IdleSince *metav1.Time `json:"idleSince,omitempty"`

	// ConnectedEnvironment tracks the connected environment details
	// +optional
	ConnectedEnvironment *ConnectedEnvironmentInfo `json:"connectedEnvironment,omitempty"`

	// SnapshotRestoreStatus tracks the progress of creating workspace from a registry snapshot
	// +optional
	SnapshotRestoreStatus *SnapshotRestoreStatus `json:"snapshotRestoreStatus,omitempty"`

	// Hash is an 8-character hash derived from owner and workspace name for DNS-safe hostnames
	// Format: hash(owner-workspaceName)
	// +optional
	Hash string `json:"hash,omitempty"`

	// Subdomain is the subdomain assigned to this workspace's workmachine (e.g., "beanbag.khost.dev")
	// +optional
	Subdomain string `json:"subdomain,omitempty"`

	// ExposedRoutes contains the URLs for user-exposed HTTP ports
	// Keys are port numbers as strings, values are the full URLs
	// Example: {"3000": "https://p3000-a1b2c3d4.example.khost.dev"}
	// +optional
	ExposedRoutes map[string]string `json:"exposedRoutes,omitempty"`

	// LastRestoredSnapshot tracks the last snapshot that was restored to this workspace
	// Used for automatic parent lineage tracking when new snapshots are created
	// +optional
	LastRestoredSnapshot *WorkspaceLastRestoredSnapshotInfo `json:"lastRestoredSnapshot,omitempty"`
}

// WorkspaceLastRestoredSnapshotInfo tracks the last restored snapshot for lineage
type WorkspaceLastRestoredSnapshotInfo struct {
	// Name is the name of the snapshot that was restored
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// RestoredAt is when the snapshot was restored
	// +kubebuilder:validation:Required
	RestoredAt metav1.Time `json:"restoredAt"`
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

// FromSnapshotRef specifies a pushed snapshot to create the workspace from
type FromSnapshotRef struct {
	// SnapshotName is the name of the snapshot resource to clone from
	// The snapshot must have status.registryStatus.pushed=true
	// +kubebuilder:validation:Required
	SnapshotName string `json:"snapshotName"`
}

// SnapshotRestorePhase represents the current phase of snapshot restoration
type SnapshotRestorePhase string

const (
	// SnapshotRestorePhasePending indicates restoration is pending to start
	SnapshotRestorePhasePending SnapshotRestorePhase = "Pending"

	// SnapshotRestorePhasePulling indicates snapshot is being pulled from registry
	SnapshotRestorePhasePulling SnapshotRestorePhase = "Pulling"

	// SnapshotRestorePhaseRestoring indicates snapshot is being restored to workspace
	SnapshotRestorePhaseRestoring SnapshotRestorePhase = "Restoring"

	// SnapshotRestorePhaseCompleted indicates restoration completed successfully
	SnapshotRestorePhaseCompleted SnapshotRestorePhase = "Completed"

	// SnapshotRestorePhaseFailed indicates restoration failed
	SnapshotRestorePhaseFailed SnapshotRestorePhase = "Failed"
)

// SnapshotRestoreStatus tracks the progress of creating workspace from a registry snapshot
type SnapshotRestoreStatus struct {
	// Phase represents the current phase of snapshot restoration
	// +kubebuilder:validation:Enum=Pending;Pulling;Restoring;Completed;Failed
	// +optional
	Phase SnapshotRestorePhase `json:"phase,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// SourceSnapshot is the name of the snapshot being restored from
	// +optional
	SourceSnapshot string `json:"sourceSnapshot,omitempty"`

	// ImageRef is the registry image reference being pulled
	// +optional
	ImageRef string `json:"imageRef,omitempty"`

	// SnapshotRequestName is the name of the SnapshotRequest created for pulling
	// +optional
	SnapshotRequestName string `json:"snapshotRequestName,omitempty"`

	// StartTime when restoration started
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime when restoration completed (success or failure)
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// ErrorMessage if restoration failed
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={kloudlite,workspaces}
// +kubebuilder:printcolumn:name="Display Name",type=string,JSONPath=`.spec.displayName`
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.spec.owner`
// +kubebuilder:printcolumn:name="WorkMachine",type=string,JSONPath=`.spec.workmachine`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.spec.status`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.status.connectedEnvironment.name`
// +kubebuilder:printcolumn:name="Connections",type=integer,JSONPath=`.status.activeConnections`
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
