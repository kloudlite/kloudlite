package v1

import (
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:JSONPath=".spec.deviceId",name="Device ID",type=string
// +kubebuilder:printcolumn:JSONPath=".status.assignedIP",name="Assigned IP",type=string
// +kubebuilder:printcolumn:JSONPath=".status.phase",name="Phase",type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastSeen",name="Last Seen",type=date

// WireGuardDevice represents a client device connecting to the VPN
type WireGuardDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WireGuardDeviceSpec   `json:"spec,omitempty"`
	Status WireGuardDeviceStatus `json:"status,omitempty"`
}

// GetStatus returns the reconciler status
func (wd *WireGuardDevice) GetStatus() *reconciler.Status {
	return &wd.Status.Status
}

// WireGuardDeviceSpec defines the desired state of WireGuardDevice
type WireGuardDeviceSpec struct {
	// DeviceID is the unique identifier for this device (UUID)
	// +kubebuilder:validation:Required
	DeviceID string `json:"deviceId"`

	// UserRef is the username of the user who owns this device
	// +kubebuilder:validation:Required
	UserRef string `json:"userRef"`

	// WorkMachineRef is the name of the WorkMachine this device connects to
	// +kubebuilder:validation:Required
	WorkMachineRef string `json:"workMachineRef"`

	// DeviceName is a human-readable name for the device (e.g., "John's MacBook")
	// +kubebuilder:validation:Optional
	DeviceName string `json:"deviceName,omitempty"`

	// Platform indicates the operating system (darwin, linux, windows)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=darwin;linux;windows
	Platform string `json:"platform,omitempty"`
}

// WireGuardDeviceStatus defines the observed state of WireGuardDevice
type WireGuardDeviceStatus struct {
	reconciler.Status `json:",inline"`

	// Phase represents the current state of the device
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Pending;Provisioning;Ready;Error
	Phase string `json:"phase,omitempty"`

	// PublicKey is the WireGuard public key for this device
	// +kubebuilder:validation:Optional
	PublicKey string `json:"publicKey,omitempty"`

	// AssignedIP is the IP address allocated to this device
	// +kubebuilder:validation:Optional
	AssignedIP string `json:"assignedIP,omitempty"`

	// LastSeen is the timestamp when this device last connected
	// +kubebuilder:validation:Optional
	LastSeen *metav1.Time `json:"lastSeen,omitempty"`

	// ConfigGeneration increments when the device configuration changes
	// +kubebuilder:validation:Optional
	ConfigGeneration int64 `json:"configGeneration,omitempty"`

	// Message provides additional information about the current state
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WireGuardDeviceList contains a list of WireGuardDevice
type WireGuardDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WireGuardDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WireGuardDevice{}, &WireGuardDeviceList{})
}
