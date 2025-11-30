package v1

import (
	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PortMapping defines mapping between service and workspace ports for intercepts
type PortMapping struct {
	// ServicePort is the port exposed by the service
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	ServicePort int32 `json:"servicePort"`

	// WorkspacePort is the port in the workspace pod
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	WorkspacePort int32 `json:"workspacePort"`

	// Protocol is the protocol used (TCP/UDP)
	// +kubebuilder:validation:Enum=TCP;UDP;SCTP
	// +kubebuilder:default=TCP
	// +optional
	Protocol corev1.Protocol `json:"protocol,omitempty"`
}

// InterceptSpec defines a service to intercept when connected to an environment
type InterceptSpec struct {
	// ServiceName in the connected environment to intercept
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	ServiceName string `json:"serviceName"`

	// PortMappings defines how service ports map to workspace ports
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	PortMappings []PortMapping `json:"portMappings"`
}

// EnvironmentConnectionSpec defines the environment connection and associated intercepts
type EnvironmentConnectionSpec struct {
	// EnvironmentRef references the environment to connect to
	// +kubebuilder:validation:Required
	EnvironmentRef corev1.ObjectReference `json:"environmentRef"`

	// Intercepts defines services to intercept in this environment
	// When environment is disconnected, all intercepts are automatically removed
	// +optional
	Intercepts []InterceptSpec `json:"intercepts,omitempty"`
}

// PackageSpec is an alias to packages.kloudlite.io/v1 PackageSpec
type PackageSpec = packagesv1.PackageSpec

// InstalledPackage is an alias to packages.kloudlite.io/v1 InstalledPackage
type InstalledPackage = packagesv1.InstalledPackage

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

	// Packages list of Nix packages to install in the workspace
	// +optional
	Packages []PackageSpec `json:"packages,omitempty"`

	// Status indicates whether the workspace is active or suspended
	// +kubebuilder:validation:Enum=active;suspended;archived
	// +kubebuilder:default=active
	Status string `json:"status,omitempty"`

	// CopyFrom specifies the source workspace name to clone directory from
	// This field is automatically cleared after successful cloning
	// +optional
	CopyFrom string `json:"copyFrom,omitempty"`

	// HttpExpose defines HTTP ports to expose from the workspace via ingress
	// Each port will get an ingress route with hostname p{port}-{hash}.{subdomain}
	// +optional
	HttpExpose []int32 `json:"httpExpose,omitempty"`
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

// InterceptStatus tracks the status of an active service intercept
type InterceptStatus struct {
	// ServiceName being intercepted
	ServiceName string `json:"serviceName"`

	// Phase of the intercept (Pending, Active, Failed, etc.)
	// +optional
	Phase string `json:"phase,omitempty"`

	// Message provides additional information about the intercept status
	// +optional
	Message string `json:"message,omitempty"`
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

	// InstalledPackages list of successfully installed packages
	// +optional
	InstalledPackages []InstalledPackage `json:"installedPackages,omitempty"`

	// FailedPackages list of packages that failed to install
	// +optional
	FailedPackages []string `json:"failedPackages,omitempty"`

	// PackageInstallationMessage provides information about package installation
	// +optional
	PackageInstallationMessage string `json:"packageInstallationMessage,omitempty"`

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

	// ActiveIntercepts tracks the status of active service intercepts
	// +optional
	ActiveIntercepts []InterceptStatus `json:"activeIntercepts,omitempty"`

	// CloningStatus tracks the progress of workspace directory cloning
	// +optional
	CloningStatus *WorkspaceCloningStatus `json:"cloningStatus,omitempty"`

	// SourceCloningStatus tracks when this workspace is being used as a cloning source
	// +optional
	SourceCloningStatus *WorkspaceSourceCloningStatus `json:"sourceCloningStatus,omitempty"`

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

// CloningPhase represents the current phase of workspace cloning
type CloningPhase string

const (
	// CloningPhasePending indicates cloning is pending to start
	CloningPhasePending CloningPhase = "Pending"

	// CloningPhaseSuspending indicates source workspace is being suspended
	CloningPhaseSuspending CloningPhase = "Suspending"

	// CloningPhaseCreatingCopyJob indicates sender and receiver jobs are being created
	CloningPhaseCreatingCopyJob CloningPhase = "CreatingCopyJob"

	// CloningPhaseWaitingForCopyCompletion indicates waiting for directory copy to complete
	CloningPhaseWaitingForCopyCompletion CloningPhase = "WaitingForCopyCompletion"

	// CloningPhaseVerifyingCopy indicates verifying the directory was copied successfully
	CloningPhaseVerifyingCopy CloningPhase = "VerifyingCopy"

	// CloningPhaseResuming indicates source workspace is being resumed
	CloningPhaseResuming CloningPhase = "Resuming"

	// CloningPhaseCompleted indicates cloning completed successfully
	CloningPhaseCompleted CloningPhase = "Completed"

	// CloningPhaseFailed indicates cloning failed
	CloningPhaseFailed CloningPhase = "Failed"
)

// DirectoryCopyJobStatus tracks the status of directory copy sender/receiver jobs
type DirectoryCopyJobStatus struct {
	// SenderJobName is the name of the sender job
	// +optional
	SenderJobName string `json:"senderJobName,omitempty"`

	// ReceiverJobName is the name of the receiver job
	// +optional
	ReceiverJobName string `json:"receiverJobName,omitempty"`

	// SenderPodIP is the IP address of the sender pod
	// +optional
	SenderPodIP string `json:"senderPodIP,omitempty"`

	// Started indicates if the copy job has started
	// +optional
	Started bool `json:"started,omitempty"`

	// Completed indicates if the copy completed successfully
	// +optional
	Completed bool `json:"completed,omitempty"`

	// Failed indicates if the copy failed
	// +optional
	Failed bool `json:"failed,omitempty"`

	// Message provides additional information about the copy status
	// +optional
	Message string `json:"message,omitempty"`

	// StartTime when the copy job started
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime when the copy job completed
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

// WorkspaceCloningStatus tracks the overall cloning progress for the target workspace
type WorkspaceCloningStatus struct {
	// Phase represents the current phase of cloning
	// +optional
	Phase CloningPhase `json:"phase,omitempty"`

	// Message provides additional information about the current cloning state
	// +optional
	Message string `json:"message,omitempty"`

	// SourceWorkspaceName is the name of the workspace being cloned from
	// +optional
	SourceWorkspaceName string `json:"sourceWorkspaceName,omitempty"`

	// SourceWorkmachineName is the WorkMachine where the source workspace is located
	// +optional
	SourceWorkmachineName string `json:"sourceWorkmachineName,omitempty"`

	// SourceFolderName is the folder name of the source workspace
	// +optional
	SourceFolderName string `json:"sourceFolderName,omitempty"`

	// CopyJobStatus tracks the directory copy job status
	// +optional
	CopyJobStatus *DirectoryCopyJobStatus `json:"copyJobStatus,omitempty"`

	// StartTime when cloning started
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime when cloning completed (success or failure)
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// ErrorMessage if cloning failed
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// SourceCloningPhase represents the phase of the source workspace during cloning
type SourceCloningPhase string

const (
	// SourceCloningPhaseSuspending indicates source is being suspended for cloning
	SourceCloningPhaseSuspending SourceCloningPhase = "Suspending"

	// SourceCloningPhaseCopying indicates source directory is being copied
	SourceCloningPhaseCopying SourceCloningPhase = "Copying"

	// SourceCloningPhaseResuming indicates source is being resumed after cloning
	SourceCloningPhaseResuming SourceCloningPhase = "Resuming"
)

// WorkspaceSourceCloningStatus tracks the status of this workspace when used as a clone source
type WorkspaceSourceCloningStatus struct {
	// Phase represents the current phase of source workspace during cloning
	// +optional
	Phase SourceCloningPhase `json:"phase,omitempty"`

	// Message provides additional information
	// +optional
	Message string `json:"message,omitempty"`

	// TargetWorkspaceName is the name of the workspace being created from this source
	// +optional
	TargetWorkspaceName string `json:"targetWorkspaceName,omitempty"`

	// TargetWorkmachineName is the WorkMachine where the target workspace will be created
	// +optional
	TargetWorkmachineName string `json:"targetWorkmachineName,omitempty"`

	// StartTime when this workspace started being used as a clone source
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime when the cloning operation completed
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
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
