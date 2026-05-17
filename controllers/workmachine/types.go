package workmachine

// InstanceState represents the current state of a cloud instance
type InstanceState string

const (
	// InstanceStatePending means the instance is being created
	InstanceStatePending InstanceState = "pending"

	// InstanceStateRunning means the instance is running
	InstanceStateRunning InstanceState = "running"

	// InstanceStateStopping means the instance is stopping
	InstanceStateStopping InstanceState = "stopping"

	// InstanceStateStopped means the instance is stopped
	InstanceStateStopped InstanceState = "stopped"

	// InstanceStateTerminating means the instance is being terminated
	InstanceStateTerminating InstanceState = "terminating"

	// InstanceStateTerminated means the instance has been terminated
	InstanceStateTerminated InstanceState = "terminated"

	// InstanceStateError means there was an error with the instance
	InstanceStateError InstanceState = "error"

	// InstanceStateNotFound means the instance doesn't exist
	InstanceStateNotFound InstanceState = "not-found"
)

// InstanceInfo contains information about a cloud instance
type InstanceInfo struct {
	// InstanceID is the cloud provider's unique identifier for the instance
	InstanceID string

	// State is the current state of the instance
	State InstanceState

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
