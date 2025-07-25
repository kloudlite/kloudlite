package domain

import (
	"github.com/kloudlite/api/apps/accounts/internal/entities"
	"github.com/kloudlite/api/apps/accounts/internal/env"
	iamT "github.com/kloudlite/api/apps/iam/types"
	authrpc "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	"golang.org/x/net/context"
	"log/slog"
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

	EnsureKloudliteRegistryCredentials(ctx UserContext, accountName string) error

	AvailableKloudliteRegions(ctx UserContext) ([]*AvailableKloudliteRegion, error)
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
	authClient    authrpc.AuthInternalClient
	iamClient     iam.IAMClient
	consoleClient console.ConsoleClient
	// containerRegistryClient container_registry.ContainerRegistryClient
	commsClient comms.CommsClient

	accountRepo    repos.DbRepo[*entities.Account]
	invitationRepo repos.DbRepo[*entities.Invitation]
	// accountInviteTokenRepo cache.Repo[*entities.Invitation]

	k8sClient k8s.Client

	Env *env.AccountsEnv

	logger *slog.Logger
}

var Module = fx.Module("domain", fx.Provide(func(
	iamCli iam.IAMClient,
	consoleClient console.ConsoleClient,
	authClient authrpc.AuthInternalClient,
	commsClient comms.CommsClient,
	k8sClient k8s.Client,
	accountRepo repos.DbRepo[*entities.Account],
	invitationRepo repos.DbRepo[*entities.Invitation],
	ev *env.AccountsEnv,
	logger *slog.Logger,
) Domain {
	return &domain{
		authClient:     authClient,
		iamClient:      iamCli,
		consoleClient:  consoleClient,
		commsClient:    commsClient,
		k8sClient:      k8sClient,
		accountRepo:    accountRepo,
		invitationRepo: invitationRepo,

		Env: ev,

		logger: logger,
	}
}))
