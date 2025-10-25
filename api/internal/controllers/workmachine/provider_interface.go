package workmachine

// import (
// 	"context"
//
// 	// "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/errors"
// 	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/types"
// 	workmachinev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
// )
//
// // CloudProviderInterface defines the contract that all cloud provider implementations must follow
// // This interface provides explicit methods for CRUD operations, keeping orchestration logic in the controller
// type CloudProviderInterface interface {
// 	// ValidatePermissions checks if the configured credentials have all required permissions
// 	// Returns nil if all permissions are available, otherwise returns detailed error
// 	ValidatePermissions(ctx context.Context) error
//
// 	// Instance Operations - explicit methods for instance lifecycle management
// 	// Controllers should handle orchestration (check-if-exists, create-if-not logic)
//
// 	// CreateInstance creates a new cloud instance for the WorkMachine
// 	// Returns instance information including ID, IPs, and state
// 	// Returns ResourceAlreadyExistsError if instance already exists
// 	CreateInstance(ctx context.Context, wm *workmachinev1.WorkMachine) (*types.InstanceInfo, error)
//
// 	// GetInstance retrieves the current status of an instance by its ID
// 	// Returns ResourceNotFoundError if instance doesn't exist
// 	GetInstance(ctx context.Context, instanceID string) (*types.InstanceInfo, error)
//
// 	// StartInstance starts a stopped instance
// 	// Returns error if instance doesn't exist or cannot be started
// 	StartInstance(ctx context.Context, instanceID string) error
//
// 	// StopInstance stops a running instance (does not terminate)
// 	// Returns error if instance doesn't exist or cannot be stopped
// 	StopInstance(ctx context.Context, instanceID string) error
//
// 	// DeleteInstance permanently deletes the instance
// 	// Returns ResourceNotFoundError if instance doesn't exist
// 	DeleteInstance(ctx context.Context, instanceID string) error
//
// 	// DNS Operations - explicit methods for DNS record management
//
// 	// UpsertDNSRecord creates or updates a DNS record
// 	// fqdn is the fully qualified domain name (e.g., "machine-1.example.com")
// 	// Controller is responsible for constructing the FQDN from WorkMachine spec
// 	UpsertDNSRecord(ctx context.Context, fqdn, publicIP string) error
//
// 	// DeleteDNSRecord deletes a DNS record
// 	// fqdn is the fully qualified domain name
// 	// Controller is responsible for constructing the FQDN from WorkMachine spec
// 	DeleteDNSRecord(ctx context.Context, fqdn, publicIP string) error
// }
//
// // NewCloudProvider creates a new cloud provider implementation based on the WorkMachine spec
// func NewCloudProvider(wm *workmachinev1.WorkMachine) (CloudProviderInterface, error) {
// 	switch wm.Spec.Provider {
// 	case workmachinev1.AWS:
// 		return nil, errors.NewProviderNotImplementedError("aws - use getOrCreateAWSProvider() in controller")
// 	case workmachinev1.GCP:
// 		return nil, errors.NewProviderNotImplementedError("gcp")
// 	case workmachinev1.Azure:
// 		return nil, errors.NewProviderNotImplementedError("azure")
// 	default:
// 		return nil, errors.NewProviderNotConfiguredError()
// 	}
// }
