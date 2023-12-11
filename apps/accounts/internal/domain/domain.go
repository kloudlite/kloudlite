package domain

import (
	"go.uber.org/fx"
	"golang.org/x/net/context"
	"kloudlite.io/apps/accounts/internal/entities"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
)

type CheckNameAvailabilityOutput struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames,omitempty"`
}

type AccountService interface {
	CheckNameAvailability(ctx context.Context, name string) (*CheckNameAvailabilityOutput, error)

	ListAccounts(ctx UserContext) ([]*entities.Account, error)
	GetAccount(ctx UserContext, name string) (*entities.Account, error)

	CreateAccount(ctx UserContext, account entities.Account) (*entities.Account, error)
	UpdateAccount(ctx UserContext, account entities.Account) (*entities.Account, error)
	DeleteAccount(ctx UserContext, name string) (bool, error)

	ResyncAccount(ctx UserContext, name string) error

	ActivateAccount(ctx UserContext, name string) (bool, error)
	DeactivateAccount(ctx UserContext, name string) (bool, error)
}

type InvitationService interface {
	InviteMembers(ctx UserContext, accountName string, invitations []*entities.Invitation) ([]*entities.Invitation, error)
	ResendInviteEmail(ctx UserContext, accountName string, invitationId repos.ID) (bool, error)

	ListInvitations(ctx UserContext, accountName string) ([]*entities.Invitation, error)
	GetInvitation(ctx UserContext, accountName string, invitationId repos.ID) (*entities.Invitation, error)

	ListInvitationsForUser(ctx UserContext, onlyPending bool) ([]*entities.Invitation, error)

	DeleteInvitation(ctx UserContext, accountName string, invitationId repos.ID) (bool, error)

	AcceptInvitation(ctx UserContext, accountName string, inviteToken string) (bool, error)
	RejectInvitation(ctx UserContext, accountName string, inviteToken string) (bool, error)
}

type MembershipService interface {
	ListMembershipsForUser(ctx UserContext) ([]*entities.AccountMembership, error)
	ListMembershipsForAccount(ctx UserContext, accountName string, role *iamT.Role) ([]*entities.AccountMembership, error)

	GetAccountMembership(ctx UserContext, accountName string) (*entities.AccountMembership, error)

	RemoveAccountMembership(ctx UserContext, accountName string, memberId repos.ID) (bool, error)
	UpdateAccountMembership(ctx UserContext, accountName string, memberId repos.ID, role iamT.Role) (bool, error)
}

type Domain interface {
	AccountService
	InvitationService
	MembershipService
}

type domain struct {
	authClient              auth.AuthClient
	iamClient               iam.IAMClient
	consoleClient           console.ConsoleClient
	containerRegistryClient container_registry.ContainerRegistryClient
	commsClient             comms.CommsClient

	accountRepo    repos.DbRepo[*entities.Account]
	invitationRepo repos.DbRepo[*entities.Invitation]
	// accountInviteTokenRepo cache.Repo[*entities.Invitation]

	k8sClient k8s.Client

	logger logging.Logger
}

func NewDomain(
	iamCli iam.IAMClient,
	consoleClient console.ConsoleClient,
	// containerRegistryClient container_registry.ContainerRegistryClient,
	authClient auth.AuthClient,
	commsClient comms.CommsClient,

	k8sClient k8s.Client,

	accountRepo repos.DbRepo[*entities.Account],
	invitationRepo repos.DbRepo[*entities.Invitation],
	// accountInviteTokenRepo cache.Repo[*entities.Invitation],

	logger logging.Logger,
) Domain {
	return &domain{
		authClient:    authClient,
		iamClient:     iamCli,
		consoleClient: consoleClient,
		commsClient:   commsClient,

		k8sClient: k8sClient,

		accountRepo:    accountRepo,
		invitationRepo: invitationRepo,
		// accountInviteTokenRepo: accountInviteTokenRepo,

		logger: logger,
	}
}

var Module = fx.Module("domain", fx.Provide(NewDomain))
