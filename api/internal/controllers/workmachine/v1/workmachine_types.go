package v1

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
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

	// +kubebuilder:default="0.0.0.0/0"
	AllowedCIDR string `json:"allowedCIDR,omitempty"`

	// Provider specifies the cloud provider (aws, gcp, azure, or empty for k8s deployment)
	// +optional
	Provider CloudProvider `json:"provider,omitempty"`

	// AWSProvider contains AWS-specific configuration
	// +optional
	AWSProvider *AWSProviderConfig `json:"aws,omitempty"`

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

// AWSProviderConfig contains AWS-specific configuration for WorkMachine
type AWSProviderConfig struct {
	// Region is the AWS region where the instance will be created
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// AvailabilityZone is the specific AZ within the region
	// +optional
	AvailabilityZone string `json:"availabilityZone,omitempty"`

	// VPC_ID is the ID of the VPC where the instance will be created
	// +kubebuilder:validation:Required
	VPC_ID string `json:"vpcID"`

	// SubnetID is the ID of the subnet where the instance will be created
	// +kubebuilder:validation:Required
	SubnetID string `json:"subnetID"`

	// AMI is the Amazon Machine Image ID to use for the instance
	// +kubebuilder:validation:Required
	AMI string `json:"ami"`

	// MachineType is the EC2 instance type (e.g., m5.large, t3.medium)
	// +kubebuilder:validation:Required
	MachineType ec2types.InstanceType `json:"machineType"`

	// VolumeSize is the size of the root EBS volume in GB
	// +kubebuilder:default=50
	// +kubebuilder:validation:Minimum=20
	// +kubebuilder:validation:Maximum=1000
	VolumeSize int32 `json:"volumeSize"`

	// VolumeType is the EBS volume type (gp3, gp2, io1, io2)
	// +kubebuilder:default=gp3
	// +kubebuilder:validation:Enum=gp3;gp2;io1;io2
	VolumeType ec2types.VolumeType `json:"volumeType"`

	DeleteVolumePostTermination bool `json:"deleteVolumePostTermination,omitempty"`

	// IAMRole is the IAM role name to attach to the instance
	// +optional
	IAMRole *string `json:"iamRole,omitempty"`

	// SecurityGroupIDs are additional security group IDs to attach
	// (WorkMachine controller will create a dedicated SG automatically)
	// +optional
	SecurityGroupIDs []string `json:"securityGroupIDs,omitempty"`

	// K3sServerURL is the URL of the K3s server to join
	// +kubebuilder:validation:Required
	K3sServerURL string `json:"k3sServerURL"`

	// K3sTokenSecret is the name of the Secret containing the K3s join token
	// +kubebuilder:validation:Required
	K3sTokenSecret string `json:"k3sTokenSecret"`

	// Route53HostedZoneID is the ID of the Route53 hosted zone for DNS records
	// +kubebuilder:validation:Required
	Route53HostedZoneID string `json:"route53HostedZoneID"`

	// DomainName is the base domain for WorkMachine DNS records
	// (e.g., "workmachines.example.com" → "<machine-name>.workmachines.example.com")
	// +kubebuilder:validation:Required
	DomainName string `json:"domainName"`
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
	reconciler.Status `json:",inline"`

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

	// SSHPublicKey is the WorkMachine's public SSH key for all workspaces
	// This key is shared across all workspaces in the WorkMachine
	// Users can copy this key to add to other systems' authorized_keys
	// The corresponding private key is stored in a Secret
	// +optional
	SSHPublicKey string `json:"sshPublicKey,omitempty"`

	// --- AWS-specific fields ---

	// MachineID is the EC2 instance ID (only for AWS provider)
	// +optional
	MachineID string `json:"instanceID,omitempty"`

	// PublicIP is the public IP address of the EC2 instance
	// +optional
	PublicIP string `json:"publicIP,omitempty"`

	// PrivateIP is the private IP address of the EC2 instance
	// +optional
	PrivateIP string `json:"privateIP,omitempty"`

	// Region is the AWS region where the instance is running
	// +optional
	Region string `json:"region,omitempty"`

	// AvailabilityZone is the AWS AZ where the instance is running
	// +optional
	AvailabilityZone string `json:"availabilityZone,omitempty"`

	// SecurityGroupID is the ID of the dedicated security group for this WorkMachine
	// +optional
	SecurityGroupID string `json:"securityGroupID,omitempty"`

	// Route53RecordSet is the DNS record for this WorkMachine
	// +optional
	Route53RecordSet string `json:"route53RecordSet,omitempty"`

	// K3sJoinStatus indicates whether the K3s agent successfully joined the cluster
	// +optional
	K3sJoinStatus string `json:"k3sJoinStatus,omitempty"`

	// --- Auto-shutdown fields ---

	// LastWorkspaceActivity is the last time any workspace was active on this WorkMachine
	// Used for auto-shutdown logic
	// +optional
	LastWorkspaceActivity *metav1.Time `json:"lastWorkspaceActivity,omitempty"`

	// ActiveWorkspaceCount is the number of active (non-suspended) workspaces
	// +optional
	ActiveWorkspaceCount int32 `json:"activeWorkspaceCount,omitempty"`
}

// // MachineState represents the current state of a cloud instance
// type MachineState string
//
// const (
// 	// MachineStatePending means the instance is being created
// 	MachineStatePending MachineState = "pending"
//
// 	// MachineStateRunning means the instance is running
// 	MachineStateRunning MachineState = "running"
//
// 	// MachineStateStopping means the instance is stopping
// 	MachineStateStopping MachineState = "stopping"
//
// 	// MachineStateStopped means the instance is stopped
// 	MachineStateStopped MachineState = "stopped"
//
// 	// MachineStateTerminating means the instance is being terminated
// 	MachineStateTerminating MachineState = "terminating"
//
// 	// MachineStateTerminated means the instance has been terminated
// 	MachineStateTerminated MachineState = "terminated"
//
// 	// MachineStateError means there was an error with the instance
// 	MachineStateError MachineState = "error"
//
// 	// MachineStateNotFound means the instance doesn't exist
// 	MachineStateNotFound MachineState = "not-found"
// )

// MachineInfo contains information about a cloud instance
type MachineInfo struct {
	// MachineID is the cloud provider's unique identifier for the instance
	MachineID string

	// State is the current state of the instance
	State MachineState

	// PublicIP is the public IP address of the instance (if available)
	PublicIP string

	// PrivateIP is the private IP address of the instance
	PrivateIP string

	// Region is the cloud region where the instance is running
	Region string

	// AvailabilityZone is the availability zone within the region
	AvailabilityZone string

	// Message provides additional information about the instance state
	Message string

	// SecurityGroupID is the ID of the security group attached to the instance
	SecurityGroupID string

	// K3sJoinStatus indicates whether the K3s agent successfully joined the cluster
	K3sJoinStatus string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkMachineList contains a list of WorkMachine
type WorkMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []WorkMachine `json:"items"`
}
