package cloud

import (
	"context"

	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
)

type Provider interface {
	// ValidatePermissions checks if the configured credentials have all required permissions
	// Returns nil if all required permissions are available, otherwise returns detailed wrapped error
	ValidatePermissions(ctx context.Context) error

	// CreateMachine creates a new cloud instance for the WorkMachine
	// Returns instance information including ID, IPs, and state
	// Returns ResourceAlreadyExistsError if instance already exists
	CreateMachine(ctx context.Context, wm *v1.WorkMachine) (*v1.MachineInfo, error)

	// GetMachineStatus retrieves the current status of an instance by its ID
	// Returns ResourceNotFoundError if instance doesn't exist
	GetMachineStatus(ctx context.Context, machineID string) (*v1.MachineInfo, error)

	// StartMachine starts a stopped instance
	StartMachine(ctx context.Context, machineID string) error

	// StopMachine stops a running instance
	StopMachine(ctx context.Context, machineID string) error

	// Reboot the Machine
	RebootMachine(ctx context.Context, machineID string) error

	// IncreaseVolumeSize increases root volume size for root volume of the machine
	IncreaseVolumeSize(ctx context.Context, machineID string, newSize int32) error

	// ChangeMachine changes the instance type of the machine
	ChangeMachine(ctx context.Context, machineID string, newInstanceType string) error

	// DeleteMachine permanently deletes the instance
	DeleteMachine(ctx context.Context, machineID string) error
}
