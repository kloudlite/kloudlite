package domain

import (
	"context"
	"github.com/kloudlite/api/apps/accounts/internal/entities"
	"github.com/kloudlite/api/apps/accounts/internal/env"
	iamT "github.com/kloudlite/api/apps/iam/types"
	authrpc "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
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

	ActivateAccount(ctx UserContext, name string) (bool, error)
	DeactivateAccount(ctx UserContext, name string) (bool, error)

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

type TeamService interface {
	CreateTeam(ctx UserContext, team entities.Team) (*entities.Team, error)
	GetTeam(ctx UserContext, teamId repos.ID) (*entities.Team, error)
	ListTeams(ctx UserContext) ([]*entities.Team, error)
	SearchTeams(ctx context.Context, query string, limit, offset int) ([]*entities.Team, int, error)
	UpdateTeam(ctx UserContext, teamId repos.ID, displayName string) (*entities.Team, error)
	CheckTeamSlugAvailability(ctx context.Context, slug string) (*CheckNameAvailabilityOutput, error)
	GenerateTeamSlugSuggestions(ctx context.Context, displayName string) []string
	GetUserRoleInTeam(ctx context.Context, userId repos.ID, teamId repos.ID) (iamT.Role, error)
}


type PlatformService interface {
	// Platform settings management
	GetPlatformSettings(ctx context.Context) (*entities.PlatformSettings, error)
	UpdatePlatformSettings(ctx UserContext, settings entities.PlatformSettings) (*entities.PlatformSettings, error)
	InitializePlatform(ctx context.Context) error
	
	
	// Team approval requests
	RequestTeamCreation(ctx UserContext, request entities.TeamApprovalRequest) (*entities.TeamApprovalRequest, error)
	ListTeamApprovalRequests(ctx UserContext, status *entities.ApprovalStatus) ([]*entities.TeamApprovalRequest, error)
	GetTeamApprovalRequest(ctx UserContext, requestId repos.ID) (*entities.TeamApprovalRequest, error)
	ApproveTeamRequest(ctx UserContext, requestId repos.ID) (*entities.Team, error)
	RejectTeamRequest(ctx UserContext, requestId repos.ID, reason string) (*entities.TeamApprovalRequest, error)
	IsTeamSlugAvailableForRequest(ctx context.Context, slug string) (bool, error)
	
	// Platform user invitations
	InvitePlatformUser(ctx UserContext, email, role string) (*entities.PlatformInvitation, error)
	ListPlatformInvitations(ctx UserContext, status *string) ([]*entities.PlatformInvitation, error)
	ResendPlatformInvitation(ctx UserContext, invitationId repos.ID) error
	CancelPlatformInvitation(ctx UserContext, invitationId repos.ID) error
	AcceptPlatformInvitation(ctx context.Context, token string) error
}

type Domain interface {
	AccountService
	InvitationService
	MembershipService
	TeamService
	PlatformService
}

type domain struct {
	authClient    authrpc.AuthInternalClient
	iamClient     iam.IAMClient
	consoleClient console.ConsoleClient

	accountRepo    repos.DbRepo[*entities.Account]
	invitationRepo repos.DbRepo[*entities.Invitation]
	teamRepo       repos.DbRepo[*entities.Team]
	teamMembershipRepo repos.DbRepo[*entities.TeamMembership]
	teamApprovalRepo repos.DbRepo[*entities.TeamApprovalRequest]
	platformSettingsRepo repos.DbRepo[*entities.PlatformSettings]
	platformInvitationRepo repos.DbRepo[*entities.PlatformInvitation]

	k8sClient k8s.Client

	Env *env.Env

	logger *slog.Logger

	*teamDomain
	*platformDomain
}

var Module = fx.Module("domain", fx.Provide(func(
	iamCli iam.IAMClient,
	consoleClient console.ConsoleClient,
	authClient authrpc.AuthInternalClient,
	k8sClient k8s.Client,
	accountRepo repos.DbRepo[*entities.Account],
	invitationRepo repos.DbRepo[*entities.Invitation],
	teamRepo repos.DbRepo[*entities.Team],
	teamMembershipRepo repos.DbRepo[*entities.TeamMembership],
	teamApprovalRepo repos.DbRepo[*entities.TeamApprovalRequest],
	platformSettingsRepo repos.DbRepo[*entities.PlatformSettings],
	platformInvitationRepo repos.DbRepo[*entities.PlatformInvitation],
	ev *env.Env,
	logger *slog.Logger,
) Domain {
	d := &domain{
		authClient:     authClient,
		iamClient:      iamCli,
		consoleClient:  consoleClient,
		k8sClient:      k8sClient,
		accountRepo:    accountRepo,
		invitationRepo: invitationRepo,
		teamRepo:       teamRepo,
		teamMembershipRepo: teamMembershipRepo,
		teamApprovalRepo: teamApprovalRepo,
		platformSettingsRepo: platformSettingsRepo,
		platformInvitationRepo: platformInvitationRepo,

		Env: ev,

		logger: logger,
	}

	// Create platform domain first
	pd := &platformDomain{
		teamApprovalRepo: teamApprovalRepo,
		platformSettingsRepo: platformSettingsRepo,
		platformInvitationRepo: platformInvitationRepo,
		teamRepo:        teamRepo,
		teamMembershipRepo: teamMembershipRepo,
		authClient:      authClient,
		logger:          logger,
		platformOwnerEmail: ev.PlatformOwnerEmail,
		webURL:          ev.WebURL,
	}
	
	// Create team domain with platform domain reference
	td := &teamDomain{
		teamRepo:       teamRepo,
		membershipRepo: teamMembershipRepo,
		platformDomain: pd,
	}
	
	d.teamDomain = td
	d.platformDomain = pd

	return d
}))
