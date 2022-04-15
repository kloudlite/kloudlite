package domain

import (
	"context"

	"kloudlite.io/apps/finance/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	EnsureAccount()
	ListAccount()
	GetAccount()
	GetStripeSetupIntent()
	CreateAccount()
	ListMemberships()
	UpdateAccount()
	UpdateBilling()
	Deactivate()
	Activate()
	InviteMember()
	UpdateMember()
	RemoveMemberships()
	AddMemberships()
	RemoveMember()
	DeleteAccount()

	// GetDevice(ctx context.Context, id repos.ID) (*entities.Device, error)
	// GetCluster(ctx context.Context, id repos.ID) (*entities.Cluster, error)
}

type InfraMessenger interface {
	// SendAddClusterAction(action entities.SetupClusterAction) error
	// SendDeleteClusterAction(action entities.DeleteClusterAction) error
	// SendUpdateClusterAction(action entities.UpdateClusterAction) error
	// SendAddDeviceAction(action entities.AddPeerAction) error
	// SendRemoveDeviceAction(entities.DeletePeerAction) error
}
