package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.spec.ownedBy`
// +kubebuilder:printcolumn:name="Machine Type",type=string,JSONPath=`.spec.machineType`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Started At",type=date,JSONPath=`.status.startedAt`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// WorkMachine represents a user's personal development machine
type WorkMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkMachineSpec   `json:"spec,omitempty"`
	Status WorkMachineStatus `json:"status,omitempty"`
}

// WorkMachineSpec defines the desired state of WorkMachine
type WorkMachineSpec struct {
	// OwnedBy is the username/email of the user who owns this machine
	// +kubebuilder:validation:Required
	OwnedBy string `json:"ownedBy"`

	// MachineType references the MachineType to use for this machine
	// +kubebuilder:validation:Required
	MachineType string `json:"machineType"`

	// TargetNamespace is the namespace where the WorkMachine workloads will run
	// Defaults to wm-{username}
	// +kubebuilder:validation:Required
	TargetNamespace string `json:"targetNamespace"`

	// DesiredState indicates whether the machine should be running or stopped
	// +kubebuilder:validation:Enum=running;stopped;disabled
	// +kubebuilder:default=stopped
	DesiredState MachineState `json:"desiredState"`

	// SSHPublicKeys for SSH access to the VM
	// +optional
	SSHPublicKeys []string `json:"sshPublicKeys,omitempty"`
}

// MachineConfiguration defines configuration options for the WorkMachine
type MachineConfiguration struct {
	// Image to use for the machine container
	// +kubebuilder:default="kloudlite/workspace:latest"
	Image string `json:"image,omitempty"`

	// Shell to use (bash, zsh, fish)
	// +kubebuilder:validation:Enum=bash;zsh;fish
	// +kubebuilder:default=bash
	Shell string `json:"shell,omitempty"`

	// DotfilesRepo to clone and apply
	// +optional
	DotfilesRepo string `json:"dotfilesRepo,omitempty"`

	// IDESettings for code-server or other IDEs
	// +optional
	IDESettings *IDESettings `json:"ideSettings,omitempty"`

	// AutoStop configuration
	// +optional
	AutoStop *AutoStopConfig `json:"autoStop,omitempty"`

	// Timezone for the machine
	// +kubebuilder:default="UTC"
	Timezone string `json:"timezone,omitempty"`
}

// IDESettings defines IDE configuration
type IDESettings struct {
	// Type of IDE (code-server, theia, jupyter)
	// +kubebuilder:validation:Enum=code-server;theia;jupyter;none
	// +kubebuilder:default=code-server
	Type string `json:"type"`

	// Extensions to install (for code-server)
	// +optional
	Extensions []string `json:"extensions,omitempty"`

	// Settings JSON (for code-server)
	// +optional
	Settings string `json:"settings,omitempty"`

	// Password for IDE access (if not using OAuth)
	// +optional
	Password string `json:"password,omitempty"`
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

// VolumeMount defines a volume to mount in the machine
type VolumeMount struct {
	// Name of the volume
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// MountPath in the container
	// +kubebuilder:validation:Required
	MountPath string `json:"mountPath"`

	// Type of volume (pvc, configmap, secret)
	// +kubebuilder:validation:Enum=pvc;configmap;secret
	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// Source name (PVC name, ConfigMap name, or Secret name)
	// +kubebuilder:validation:Required
	Source string `json:"source"`

	// ReadOnly mount
	// +kubebuilder:default=false
	ReadOnly bool `json:"readOnly,omitempty"`
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
	MachineStateError MachineState = "error"

	// MachineStateDisabled means the machine is disabled (user inactive)
	MachineStateDisabled MachineState = "disabled"
)

// WorkMachineStatus defines the observed state of WorkMachine
type WorkMachineStatus struct {
	// State is the current state of the machine
	State MachineState `json:"state,omitempty"`

	// Message provides human-readable information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// PodName is the name of the pod running this machine
	// +optional
	PodName string `json:"podName,omitempty"`

	// PodIP is the IP address of the pod
	// +optional
	PodIP string `json:"podIP,omitempty"`

	// NodeName where the pod is running
	// +optional
	NodeName string `json:"nodeName,omitempty"`

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

	// Conditions represent the latest available observations
	// +optional
	Conditions []WorkMachineCondition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// WorkMachineCondition represents a condition of the WorkMachine
type WorkMachineCondition struct {
	// Type of condition
	Type WorkMachineConditionType `json:"type"`

	// Status of the condition (True, False, Unknown)
	Status metav1.ConditionStatus `json:"status"`

	// LastTransitionTime is the last time the condition changed
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a unique, one-word, CamelCase reason for the condition's last transition
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable message about the last transition
	// +optional
	Message string `json:"message,omitempty"`
}

// WorkMachineConditionType represents types of WorkMachine conditions
type WorkMachineConditionType string

const (
	// WorkMachineConditionReady indicates the machine is ready
	WorkMachineConditionReady WorkMachineConditionType = "Ready"

	// WorkMachineConditionPodCreated indicates the pod has been created
	WorkMachineConditionPodCreated WorkMachineConditionType = "PodCreated"

	// WorkMachineConditionAccessible indicates the machine is accessible
	WorkMachineConditionAccessible WorkMachineConditionType = "Accessible"

	// WorkMachineConditionAutoStopScheduled indicates auto-stop is scheduled
	WorkMachineConditionAutoStopScheduled WorkMachineConditionType = "AutoStopScheduled"

	// WorkMachineConditionDeletionBlocked indicates deletion is blocked due to active resources
	WorkMachineConditionDeletionBlocked WorkMachineConditionType = "DeletionBlocked"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkMachineList contains a list of WorkMachine
type WorkMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []WorkMachine `json:"items"`
}
