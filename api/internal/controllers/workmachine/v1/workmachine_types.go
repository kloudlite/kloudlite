package v1

import (
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.spec.ownedBy`
// +kubebuilder:printcolumn:name="Machine Type",type=string,JSONPath=`.spec.machineType`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Started At",type=date,JSONPath=`.status.startedAt`
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// WorkMachine represents a user's personal development machine
type WorkMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkMachineSpec   `json:"spec,omitempty"`
	Status WorkMachineStatus `json:"status,omitempty"`
}

func (wm *WorkMachine) GetStatus() *reconciler.Status {
	return &wm.Status.Status
}

// WorkMachineSpec defines the desired state of WorkMachine
type WorkMachineSpec struct {
	// DisplayName is the human-readable name for the work machine
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=100
	DisplayName string `json:"displayName"`

	// OwnedBy is the username/email of the user who owns this machine
	// +kubebuilder:validation:Required
	OwnedBy string `json:"ownedBy"`

	// TargetNamespace is the namespace where the WorkMachine workloads will run
	// Defaults to wm-{username}
	// +kubebuilder:validation:Required
	TargetNamespace string `json:"targetNamespace"`

	// State indicates whether the machine should be running or stopped
	// +kubebuilder:validation:Enum=running;stopped;disabled
	// +kubebuilder:default=running
	State MachineState `json:"state"`

	// SSHPublicKeys for SSH access to the VM
	// +optional
	SSHPublicKeys []string `json:"sshPublicKeys,omitempty"`

	// MachineType is the EC2 instance type (e.g., m5.large, t3.medium)
	// +kubebuilder:validation:Required
	MachineType string `json:"machineType"`

	// VolumeSize is the size of the root EBS volume in GB
	//+kubebuilder:default=100
	//+kubebuilder:validation:Minimum=100
	//+kubebuilder:validation:Maximum=1000
	VolumeSize *int32 `json:"volumeSize"`

	// VolumeType is the volume type
	// eg. for AWS (gp3, gp2, io1, io2)
	VolumeType string `json:"volumeType,omitempty"`

	// DeleteVolumePostTermination controls whether root volume is cleaned up post deletion
	// +kubebuilder:default=true
	DeleteVolumePostTermination bool `json:"deleteVolumePostTermination,omitempty"`

	// AutoShutdown configures automatic instance shutdown when idle
	// Only applicable for cloud providers (AWS, GCP, Azure)
	// +optional
	AutoShutdown *AutoShutdownConfig `json:"autoShutdown,omitempty"`
}

type CloudProvider string

const (
	AWS   CloudProvider = "aws"
	GCP   CloudProvider = "gcp"
	Azure CloudProvider = "azure"
)

// MachineConfiguration defines configuration options for the WorkMachine
type MachineConfiguration struct {
	// AutoStop configuration
	// +optional
	AutoStop *AutoStopConfig `json:"autoStop,omitempty"`

	// Timezone for the machine
	// +kubebuilder:default="UTC"
	Timezone string `json:"timezone,omitempty"`
}

// AutoStopConfig defines auto-stop behavior
type AutoStopConfig struct {
	// Enabled determines if auto-stop is active
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`

	// IdleMinutes before stopping the machine
	// +kubebuilder:default=30
	// +kubebuilder:validation:Minimum=5
	// +kubebuilder:validation:Maximum=1440
	IdleMinutes int32 `json:"idleMinutes"`
}

// AutoShutdownConfig configures automatic EC2 instance shutdown when idle
// for cost optimization
type AutoShutdownConfig struct {
	// Enabled determines if auto-shutdown is active
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`

	// IdleThresholdMinutes is how long to wait after all workspaces are suspended
	// before shutting down the WorkMachine EC2 instance
	// +kubebuilder:default=30
	// +kubebuilder:validation:Minimum=5
	// +kubebuilder:validation:Maximum=1440
	IdleThresholdMinutes int32 `json:"idleThresholdMinutes"`

	// CheckIntervalMinutes is how often to check workspace activity
	// +kubebuilder:default=5
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=60
	CheckIntervalMinutes int32 `json:"checkIntervalMinutes"`
}

// MachineState represents the state of a WorkMachine
type MachineState string

const (
	// MachineStateRunning means the machine is running
	MachineStateRunning MachineState = "running"

	// MachineStateStopped means the machine is stopped
	MachineStateStopped MachineState = "stopped"

	// MachineStateStarting means the machine is starting up
	MachineStateStarting MachineState = "starting"

	// MachineStateStopping means the machine is stopping
	MachineStateStopping MachineState = "stopping"

	// MachineStateError means there was an error
	MachineStateErrored MachineState = "errored"

	// MachineStateDisabled means the machine is disabled (user inactive)
	MachineStateDisabled MachineState = "disabled"
)

// GPUInfo contains detailed information about GPU hardware
type GPUInfo struct {
	// HasGPU indicates whether GPU hardware is available
	// +optional
	HasGPU bool `json:"hasGPU,omitempty"`

	// Count is the number of GPUs
	// +optional
	Count int `json:"count,omitempty"`

	// Product is the GPU product name (e.g., "tesla-t4")
	// +optional
	Product string `json:"product,omitempty"`

	// DriverVersion is the NVIDIA driver version
	// +optional
	DriverVersion string `json:"driverVersion,omitempty"`

	// RuntimeConfigured indicates whether NVIDIA Container Runtime is configured
	// +optional
	RuntimeConfigured bool `json:"runtimeConfigured,omitempty"`

	// DriverInstallationStatus tracks the status of driver installation
	// Values: "not-installed", "installing", "installed", "awaiting-reboot", "ready", "error"
	// +optional
	DriverInstallationStatus string `json:"driverInstallationStatus,omitempty"`

	// DriverInstallationMessage provides detailed status about driver installation
	// +optional
	DriverInstallationMessage string `json:"driverInstallationMessage,omitempty"`
}

// WorkMachineStatus defines the observed state of WorkMachine
type WorkMachineStatus struct {
	reconciler.Status `json:",inline"`

	MachineInfo `json:",inline"`

	// StartedAt timestamp when the machine was last started
	// +optional
	StartedAt *metav1.Time `json:"startedAt,omitempty"`

	// StoppedAt timestamp when the machine was last stopped
	// +optional
	StoppedAt *metav1.Time `json:"stoppedAt,omitempty"`

	// LastActivityAt timestamp of last user activity (for auto-stop)
	// +optional
	LastActivityAt *metav1.Time `json:"lastActivityAt,omitempty"`

	// AccessURL for accessing the machine (IDE, SSH, etc.)
	// +optional
	AccessURL string `json:"accessURL,omitempty"`

	// Resources actually allocated to the machine
	// +optional
	AllocatedResources *MachineResources `json:"allocatedResources,omitempty"`

	// GPU contains detailed GPU information if hardware is present
	// +optional
	GPU *GPUInfo `json:"gpu,omitempty"`

	// SSHPublicKey is the WorkMachine's public SSH key for all workspaces
	// This key is shared across all workspaces in the WorkMachine
	// Users can copy this key to add to other systems' authorized_keys
	// The corresponding private key is stored in a Secret
	// +optional
	SSHPublicKey string `json:"sshPublicKey,omitempty"`

	// --- Auto-shutdown fields ---

	// LastWorkspaceActivity is the last time any workspace was active on this WorkMachine
	// Used for auto-shutdown logic
	// +optional
	LastWorkspaceActivity *metav1.Time `json:"lastWorkspaceActivity,omitempty"`

	// ActiveWorkspaceCount is the number of active (non-suspended) workspaces
	// +optional
	ActiveWorkspaceCount int32 `json:"activeWorkspaceCount,omitempty"`

	// IsAutoStopped when set means machine was auto-stopped by kloudlite
	IsAutoStopped bool `json:"isAutoStopped,omitempty"`

	NodeLabels     map[string]string   `json:"nodeLabels,omitempty"`
	PodTolerations []corev1.Toleration `json:"podTolerations,omitempty"`

	// --- Machine type change tracking ---

	// CurrentMachineType is the actual instance type currently running
	// Used to detect when spec.machineType changes and trigger instance type change
	// +optional
	CurrentMachineType string `json:"currentMachineType,omitempty"`

	// MachineTypeChanging indicates a machine type change is in progress
	// +optional
	MachineTypeChanging bool `json:"machineTypeChanging,omitempty"`

	// MachineTypeChangeMessage provides status updates during machine type change
	// +optional
	MachineTypeChangeMessage string `json:"machineTypeChangeMessage,omitempty"`
}

// MachineInfo contains information about a cloud instance
type MachineInfo struct {
	// MachineID is the cloud provider's unique identifier for the instance
	MachineID string `json:"machineID,omitempty"`

	// State is the current state of the instance
	State MachineState `json:"state,omitempty"`

	// RootVolumeSize is size in GBs for the root volume.
	// It is used while processing request for increasing volume size
	RootVolumeSize int32 `json:"rootVolumeSize,omitempty"`

	// PublicIP is the public IP address of the instance (if available)
	PublicIP string `json:"publicIP,omitempty"`

	// PrivateIP is the private IP address of the instance
	PrivateIP string `json:"privateIP,omitempty"`

	// Region is the cloud region where the instance is running
	Region string `json:"region,omitempty"`

	// AvailabilityZone is the availability zone within the region
	AvailabilityZone string `json:"availabilityZone,omitempty"`

	// Message provides additional information about the instance state
	Message string `json:"message,omitempty"`

	// HasGPU indicates if this machine has a GPU (stored in status for quick filtering)
	// +optional
	HasGPU bool `json:"hasGPU,omitempty"`

	// GPUModel is the GPU model name if available (e.g., "Tesla T4", "A100")
	// This is static info stored in status, real-time metrics available via metrics endpoint
	// +optional
	GPUModel string `json:"gpuModel,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkMachineList contains a list of WorkMachine
type WorkMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []WorkMachine `json:"items"`
}
