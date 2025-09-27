package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kloudlite/api/apps/accounts/internal/entities"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	authrpc "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	"go.mongodb.org/mongo-driver/mongo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
)

type platformDomain struct {
	teamApprovalRepo       repos.DbRepo[*entities.TeamApprovalRequest]
	platformSettingsRepo   repos.DbRepo[*entities.PlatformSettings]
	platformInvitationRepo repos.DbRepo[*entities.PlatformInvitation]
	teamRepo              repos.DbRepo[*entities.Team]
	teamMembershipRepo     repos.DbRepo[*entities.TeamMembership]
	authClient            authrpc.AuthInternalClient
	logger                *slog.Logger
	platformOwnerEmail    string
	webURL                string
}

// Helper method to check platform role using auth service
func (p *platformDomain) checkPlatformRole(ctx context.Context, userId string) (role string, canCreateTeams bool, canManagePlatform bool, err error) {
	resp, err := p.authClient.GetPlatformUser(ctx, &authrpc.GetPlatformUserRequest{
		UserId: userId,
	})
	if err != nil {
		// If no platform user found, return empty role
		if errors.Is(err, repos.ErrNoDocuments) {
			return "", false, false, nil
		}
		return "", false, false, err
	}

	if resp.PlatformUser == nil {
		return "", false, false, nil
	}

	role = resp.PlatformUser.Role
	canCreateTeams = role == "admin" || role == "super_admin"
	canManagePlatform = role == "admin" || role == "super_admin"

	return role, canCreateTeams, canManagePlatform, nil
}

// Platform settings management
func (p *platformDomain) GetPlatformSettings(ctx context.Context) (*entities.PlatformSettings, error) {
	settings, err := p.platformSettingsRepo.FindById(ctx, "platform-settings")
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) || errors.Is(err, repos.ErrNoDocuments) {
			// Return default settings if not found
			defaultSettings := entities.DefaultPlatformSettings()
			defaultSettings.PlatformOwnerEmail = p.platformOwnerEmail
			return defaultSettings, nil
		}
		return nil, err
	}
	return settings, nil
}

func (p *platformDomain) UpdatePlatformSettings(ctx UserContext, settings entities.PlatformSettings) (*entities.PlatformSettings, error) {
	// Get existing settings to preserve secrets
	existingSettings, err := p.platformSettingsRepo.FindById(ctx, "platform-settings")
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) && !errors.Is(err, repos.ErrNoDocuments) {
		return nil, fmt.Errorf("failed to get existing settings: %w", err)
	}
	
	// Ensure ID is always "platform-settings"
	settings.Id = "platform-settings"
	settings.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	
	// Preserve existing secrets if not provided
	if existingSettings != nil {
		// OAuth secrets
		if settings.OAuthProviders.Google.ClientSecret == "" && existingSettings.OAuthProviders.Google.ClientSecret != "" {
			settings.OAuthProviders.Google.ClientSecret = existingSettings.OAuthProviders.Google.ClientSecret
		}
		if settings.OAuthProviders.GitHub.ClientSecret == "" && existingSettings.OAuthProviders.GitHub.ClientSecret != "" {
			settings.OAuthProviders.GitHub.ClientSecret = existingSettings.OAuthProviders.GitHub.ClientSecret
		}
		if settings.OAuthProviders.Microsoft.ClientSecret == "" && existingSettings.OAuthProviders.Microsoft.ClientSecret != "" {
			settings.OAuthProviders.Microsoft.ClientSecret = existingSettings.OAuthProviders.Microsoft.ClientSecret
		}
		
		// Preserve entire cloud provider if not provided
		if settings.CloudProvider.Provider == "" && existingSettings.CloudProvider.Provider != "" {
			settings.CloudProvider = existingSettings.CloudProvider
		} else {
			// Cloud provider validation
			if existingSettings.CloudProvider.Provider != "" && settings.CloudProvider.Provider != existingSettings.CloudProvider.Provider {
				// Check if any teams exist
				teamsCount, err := p.teamRepo.Count(ctx, repos.Filter{})
				if err != nil {
					return nil, fmt.Errorf("failed to count teams: %w", err)
				}
				if teamsCount > 0 {
					return nil, errors.New("cannot change cloud provider when teams exist. Delete all teams first")
				}
			}
			
			// Preserve cloud provider secrets if not provided
			if settings.CloudProvider.Provider == existingSettings.CloudProvider.Provider {
				if settings.CloudProvider.AWS.SecretAccessKey == "" && existingSettings.CloudProvider.AWS.SecretAccessKey != "" {
					settings.CloudProvider.AWS.SecretAccessKey = existingSettings.CloudProvider.AWS.SecretAccessKey
				}
				if settings.CloudProvider.GCP.ServiceAccountKey == "" && existingSettings.CloudProvider.GCP.ServiceAccountKey != "" {
					settings.CloudProvider.GCP.ServiceAccountKey = existingSettings.CloudProvider.GCP.ServiceAccountKey
				}
				if settings.CloudProvider.Azure.ClientSecret == "" && existingSettings.CloudProvider.Azure.ClientSecret != "" {
					settings.CloudProvider.Azure.ClientSecret = existingSettings.CloudProvider.Azure.ClientSecret
				}
				if settings.CloudProvider.DigitalOcean.Token == "" && existingSettings.CloudProvider.DigitalOcean.Token != "" {
					settings.CloudProvider.DigitalOcean.Token = existingSettings.CloudProvider.DigitalOcean.Token
				}
			}
		}
	}

	// Upsert the settings
	updatedSettings, err := p.platformSettingsRepo.UpdateById(ctx, settings.Id, &settings)
	if err != nil {
		return nil, err
	}

	return updatedSettings, nil
}

func (p *platformDomain) InitializePlatform(ctx context.Context) error {
	// Check if platform is already initialized
	existingSettings, err := p.platformSettingsRepo.FindById(ctx, "platform-settings")
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) && !errors.Is(err, repos.ErrNoDocuments) {
		p.logger.Error("error checking existing platform settings", "error", err)
		return err
	}

	if existingSettings != nil && existingSettings.IsInitialized {
		// Already initialized, not an error
		p.logger.Info("platform settings already initialized")
		return nil
	}

	// Create default platform settings
	settings := entities.DefaultPlatformSettings()
	settings.IsInitialized = true
	settings.PlatformOwnerEmail = p.platformOwnerEmail
	settings.CreatedBy = common.CreatedOrUpdatedByKloudlite
	settings.LastUpdatedBy = common.CreatedOrUpdatedByKloudlite

	// Create the settings
	_, err = p.platformSettingsRepo.Create(ctx, settings)
	if err != nil {
		return err
	}

	p.logger.Info("platform settings initialized successfully")
	
	return nil
}

// Team approval requests
func (p *platformDomain) RequestTeamCreation(ctx UserContext, request entities.TeamApprovalRequest) (*entities.TeamApprovalRequest, error) {
	// Check if team slug is already taken
	existingTeam, _ := p.teamRepo.FindOne(ctx, repos.Filter{
		"slug": request.TeamSlug,
	})
	if existingTeam != nil {
		return nil, errors.Newf("team with slug %s already exists", request.TeamSlug)
	}

	// Check if there's already a pending request for this slug from any user
	existingRequest, _ := p.teamApprovalRepo.FindOne(ctx, repos.Filter{
		"teamSlug": request.TeamSlug,
		"status":   entities.ApprovalStatusPending,
	})
	if existingRequest != nil {
		return nil, errors.Newf("a pending request for team slug %s already exists", request.TeamSlug)
	}

	request.Id = p.teamApprovalRepo.NewId()
	request.RequestedBy = ctx.UserId
	request.RequestedByEmail = ctx.UserEmail
	request.RequestedAt = time.Now()
	request.Status = entities.ApprovalStatusPending
	request.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	request.LastUpdatedBy = request.CreatedBy

	createdRequest, err := p.teamApprovalRepo.Create(ctx, &request)
	if err != nil {
		return nil, err
	}

	// Send notification to platform admins
	_, err = p.authClient.CreateNotification(ctx, &authrpc.CreateNotificationRequest{
		Title:       "New Team Creation Request",
		Description: fmt.Sprintf("%s has requested to create a new team: %s", ctx.UserName, request.TeamSlug),
		Type:        "team_request",
		Target: &authrpc.NotificationTargetInternal{
			Type:            "platform_role",
			MinPlatformRole: "admin",
		},
		ActionRequired: true,
		Actions: []*authrpc.NotificationActionInternal{
			{
				Id:       "approve",
				Label:    "Approve",
				Style:    "primary",
				Endpoint: "/api/teams/approve",
				Method:   "POST",
				Data: map[string]string{
					"requestId": string(createdRequest.Id),
				},
			},
			{
				Id:       "reject",
				Label:    "Reject",
				Style:    "danger",
				Endpoint: "/api/teams/reject",
				Method:   "POST",
				Data: map[string]string{
					"requestId": string(createdRequest.Id),
				},
			},
		},
		DedupeKey: fmt.Sprintf("team-request-%s", createdRequest.Id),
	})
	if err != nil {
		p.logger.Error("failed to create notification for team request", "error", err)
		// Don't fail the request creation if notification fails
	}

	return createdRequest, nil
}

func (p *platformDomain) ListTeamApprovalRequests(ctx UserContext, status *entities.ApprovalStatus) ([]*entities.TeamApprovalRequest, error) {
	// Check if user is admin using auth service
	_, _, canManagePlatform, err := p.checkPlatformRole(ctx.Context, string(ctx.UserId))
	if err != nil {
		return nil, fmt.Errorf("failed to get platform role: %w", err)
	}

	filter := repos.Filter{}
	
	// If not an admin, only show their own requests
	if !canManagePlatform {
		filter["requestedBy"] = ctx.UserId
	}
	// Admins can see all requests

	if status != nil {
		filter["status"] = *status
	}

	requests, err := p.teamApprovalRepo.Find(ctx, repos.Query{
		Filter: filter,
		Sort: map[string]interface{}{
			"requestedAt": -1,
		},
	})
	if err != nil {
		return nil, err
	}

	return requests, nil
}

func (p *platformDomain) GetTeamApprovalRequest(ctx UserContext, requestId repos.ID) (*entities.TeamApprovalRequest, error) {
	request, err := p.teamApprovalRepo.FindById(ctx, requestId)
	if err != nil {
		return nil, err
	}

	// Check if user is admin using auth service
	_, _, canManagePlatform, err := p.checkPlatformRole(ctx.Context, string(ctx.UserId))
	if err != nil {
		return nil, fmt.Errorf("failed to get platform role: %w", err)
	}

	// Only allow access if user is admin or the requester
	if !canManagePlatform && request.RequestedBy != ctx.UserId {
		return nil, errors.New("unauthorized to view this request")
	}

	return request, nil
}

func (p *platformDomain) ApproveTeamRequest(ctx UserContext, requestId repos.ID) (*entities.Team, error) {
	// Check if user is admin using auth service
	_, _, canManagePlatform, err := p.checkPlatformRole(ctx.Context, string(ctx.UserId))
	if err != nil {
		return nil, fmt.Errorf("failed to get platform role: %w", err)
	}

	if !canManagePlatform {
		return nil, errors.New("only platform admins can approve team requests")
	}

	request, err := p.teamApprovalRepo.FindById(ctx, requestId)
	if err != nil {
		return nil, fmt.Errorf("failed to find team request: %w", err)
	}

	if request.Status != entities.ApprovalStatusPending {
		return nil, errors.Newf("request is not pending (current status: %s)", request.Status)
	}

	// Re-check if slug is still available (in case another team was created after the request)
	existingTeam, err := p.teamRepo.FindOne(ctx.Context, repos.Filter{
		"slug": request.TeamSlug,
	})
	if err != nil && !errors.Is(err, repos.ErrNoDocuments) {
		return nil, fmt.Errorf("failed to check team slug availability: %w", err)
	}
	if existingTeam != nil {
		// Mark the request as rejected due to slug conflict
		request.Status = entities.ApprovalStatusRejected
		reason := "Team slug is no longer available"
		request.RejectionReason = &reason
		reviewedBy := ctx.UserId
		request.ReviewedBy = &reviewedBy
		email := ctx.UserEmail
		request.ReviewedByEmail = &email
		now := time.Now()
		request.ReviewedAt = &now
		request.LastUpdatedBy = common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		}
		_, updateErr := p.teamApprovalRepo.UpdateById(ctx, request.Id, request)
		if updateErr != nil {
			return nil, fmt.Errorf("failed to update request status: %w", updateErr)
		}
		return nil, errors.Newf("team slug '%s' is no longer available", request.TeamSlug)
	}

	// Create the team
	newTeam := entities.Team{
		ObjectMeta: metav1.ObjectMeta{
			Name: request.TeamSlug,
		},
		ResourceMetadata: common.ResourceMetadata{
			DisplayName: request.DisplayName,
		},
		Slug:        request.TeamSlug,
		Description: request.TeamDescription,
		Region:      request.TeamRegion,
		OwnerId:     request.RequestedBy,
	}
	newTeam.Id = p.teamRepo.NewId()
	newTeam.IncrementRecordVersion()

	team, err := p.teamRepo.Create(ctx, &newTeam)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Create membership for the requester as owner
	membership := &entities.TeamMembership{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", team.Id, request.RequestedBy),
		},
		TeamId: team.Id,
		UserId: request.RequestedBy,
		Role:   iamT.RoleAccountOwner,
	}
	membership.Id = p.teamMembershipRepo.NewId()
	membership.IncrementRecordVersion()
	
	_, err = p.teamMembershipRepo.Create(ctx, membership)
	if err != nil {
		return nil, fmt.Errorf("failed to create team membership: %w", err)
	}

	// Update the request status
	request.Status = entities.ApprovalStatusApproved
	request.ReviewedBy = &ctx.UserId
	request.ReviewedByEmail = &ctx.UserEmail
	now := time.Now()
	request.ReviewedAt = &now
	request.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	_, err = p.teamApprovalRepo.UpdateById(ctx, request.Id, request)
	if err != nil {
		return nil, fmt.Errorf("failed to update request status: %w", err)
	}

	// Send notification to requester
	_, err = p.authClient.CreateNotification(ctx, &authrpc.CreateNotificationRequest{
		Title:       "Team Request Approved",
		Description: fmt.Sprintf("Your request to create team '%s' has been approved! The team has been created.", request.TeamSlug),
		Type:        "team_approved",
		Target: &authrpc.NotificationTargetInternal{
			Type:   "user",
			UserId: string(request.RequestedBy),
		},
		ActionRequired: false,
		Actions: []*authrpc.NotificationActionInternal{
			{
				Id:       "view_team",
				Label:    "View Team",
				Style:    "primary",
				Endpoint: fmt.Sprintf("/teams/%s", team.Slug),
				Method:   "GET",
			},
		},
		DedupeKey: fmt.Sprintf("team-approved-%s", request.Id),
	})
	if err != nil {
		p.logger.Error("failed to create notification for team approval", "error", err)
		// Don't fail the approval if notification fails
	}

	// Also mark the original request notification as action taken for all admins
	// This is handled automatically by the notification system when users take action

	return team, nil
}

func (p *platformDomain) RejectTeamRequest(ctx UserContext, requestId repos.ID, reason string) (*entities.TeamApprovalRequest, error) {
	// Check if user is admin using auth service
	_, _, canManagePlatform, err := p.checkPlatformRole(ctx.Context, string(ctx.UserId))
	if err != nil {
		return nil, fmt.Errorf("failed to get platform role: %w", err)
	}

	if !canManagePlatform {
		return nil, errors.New("only platform admins can reject team requests")
	}

	request, err := p.teamApprovalRepo.FindById(ctx, requestId)
	if err != nil {
		return nil, fmt.Errorf("failed to find team request: %w", err)
	}

	if request.Status != entities.ApprovalStatusPending {
		return nil, errors.Newf("request is not pending (current status: %s)", request.Status)
	}

	// Update the request status
	request.Status = entities.ApprovalStatusRejected
	request.ReviewedBy = &ctx.UserId
	request.ReviewedByEmail = &ctx.UserEmail
	now := time.Now()
	request.ReviewedAt = &now
	request.RejectionReason = &reason
	request.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	updatedRequest, err := p.teamApprovalRepo.UpdateById(ctx, request.Id, request)
	if err != nil {
		return nil, fmt.Errorf("failed to update request status: %w", err)
	}

	// Send notification to requester
	_, err = p.authClient.CreateNotification(ctx, &authrpc.CreateNotificationRequest{
		Title:       "Team Request Rejected",
		Description: fmt.Sprintf("Your request to create team '%s' has been rejected. Reason: %s", request.TeamSlug, reason),
		Type:        "team_rejected",
		Target: &authrpc.NotificationTargetInternal{
			Type:   "user",
			UserId: string(request.RequestedBy),
		},
		ActionRequired: false,
		DedupeKey:      fmt.Sprintf("team-rejected-%s", request.Id),
	})
	if err != nil {
		p.logger.Error("failed to create notification for team rejection", "error", err)
		// Don't fail the rejection if notification fails
	}

	return updatedRequest, nil
}

// Platform user invitation methods
func (p *platformDomain) InvitePlatformUser(ctx UserContext, email, role string) (*entities.PlatformInvitation, error) {
	// Check if user is super admin or admin
	userRole, _, canManagePlatform, err := p.checkPlatformRole(ctx.Context, string(ctx.UserId))
	if err != nil {
		return nil, fmt.Errorf("failed to get platform role: %w", err)
	}

	if !canManagePlatform {
		return nil, errors.New("only platform admins can invite users")
	}

	// Super admins can invite anyone, admins can invite users and other admins but not super admins
	if userRole != "super_admin" && role == "super_admin" {
		return nil, errors.New("only super admins can invite other super admins")
	}

	// Check if there's already a pending invitation for this email
	existingInvitation, _ := p.platformInvitationRepo.FindOne(ctx, repos.Filter{
		"email":  email,
		"status": "pending",
	})
	if existingInvitation != nil {
		return nil, errors.Newf("a pending invitation for %s already exists", email)
	}

	// Check if user already exists in the platform
	userResp, err := p.authClient.EnsureUserByEmail(ctx.Context, &authrpc.GetUserByEmailRequest{
		Email: email,
	})
	if err == nil && userResp.UserId != "" {
		// User exists, check if they already have a platform role
		_, _, _, err = p.checkPlatformRole(ctx.Context, userResp.UserId)
		if err == nil {
			return nil, errors.Newf("user %s already has a platform role", email)
		}
	}

	// Generate invitation token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Create invitation
	invitation := &entities.PlatformInvitation{
		Email:          email,
		Role:           role,
		InvitedBy:      string(ctx.UserId),
		InvitedByEmail: ctx.UserEmail,
		Status:         "pending",
		Token:          token,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
	}
	invitation.Id = p.platformInvitationRepo.NewId()
	invitation.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	invitation.LastUpdatedBy = invitation.CreatedBy

	createdInvitation, err := p.platformInvitationRepo.Create(ctx, invitation)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Send invitation email via auth service
	inviteLink := fmt.Sprintf("%s/auth/platform-invite?token=%s", p.webURL, token)
	
	_, err = p.authClient.SendPlatformInviteEmail(ctx.Context, &authrpc.SendPlatformInviteEmailRequest{
		Email:      email,
		Name:       email, // We don't have name, use email
		InvitedBy:  ctx.UserEmail,
		Role:       string(role),
		InviteLink: inviteLink,
	})
	if err != nil {
		p.logger.Error("failed to send invitation email", "error", err, "email", email)
		// Don't fail the invitation creation if email fails
	}

	return createdInvitation, nil
}

func (p *platformDomain) ListPlatformInvitations(ctx UserContext, status *string) ([]*entities.PlatformInvitation, error) {
	// Check if user is admin
	_, _, canManagePlatform, err := p.checkPlatformRole(ctx.Context, string(ctx.UserId))
	if err != nil {
		return nil, fmt.Errorf("failed to get platform role: %w", err)
	}

	if !canManagePlatform {
		return nil, errors.New("only platform admins can list invitations")
	}

	filter := repos.Filter{}
	if status != nil && *status != "" {
		filter["status"] = *status
	}

	invitations, err := p.platformInvitationRepo.Find(ctx, repos.Query{
		Filter: filter,
		Sort: map[string]interface{}{
			"createdAt": -1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}

	return invitations, nil
}

func (p *platformDomain) ResendPlatformInvitation(ctx UserContext, invitationId repos.ID) error {
	// Check if user is admin
	_, _, canManagePlatform, err := p.checkPlatformRole(ctx.Context, string(ctx.UserId))
	if err != nil {
		return fmt.Errorf("failed to get platform role: %w", err)
	}

	if !canManagePlatform {
		return errors.New("only platform admins can resend invitations")
	}

	invitation, err := p.platformInvitationRepo.FindById(ctx, invitationId)
	if err != nil {
		return fmt.Errorf("failed to find invitation: %w", err)
	}

	if invitation.Status != "pending" {
		return errors.New("only pending invitations can be resent")
	}

	if time.Now().After(invitation.ExpiresAt) {
		// Update status to expired
		invitation.Status = "expired"
		_, err = p.platformInvitationRepo.UpdateById(ctx, invitation.Id, invitation)
		if err != nil {
			p.logger.Error("failed to update expired invitation", "error", err)
		}
		return errors.New("invitation has expired")
	}

	// Resend invitation email
	inviteLink := fmt.Sprintf("%s/auth/platform-invite?token=%s", p.webURL, invitation.Token)
	
	_, err = p.authClient.SendPlatformInviteEmail(ctx.Context, &authrpc.SendPlatformInviteEmailRequest{
		Email:      invitation.Email,
		Name:       invitation.Email,
		InvitedBy:  ctx.UserEmail,
		Role:       invitation.Role,
		InviteLink: inviteLink,
	})
	if err != nil {
		return fmt.Errorf("failed to resend invitation email: %w", err)
	}

	return nil
}

func (p *platformDomain) CancelPlatformInvitation(ctx UserContext, invitationId repos.ID) error {
	// Check if user is admin
	_, _, canManagePlatform, err := p.checkPlatformRole(ctx.Context, string(ctx.UserId))
	if err != nil {
		return fmt.Errorf("failed to get platform role: %w", err)
	}

	if !canManagePlatform {
		return errors.New("only platform admins can cancel invitations")
	}

	invitation, err := p.platformInvitationRepo.FindById(ctx, invitationId)
	if err != nil {
		return fmt.Errorf("failed to find invitation: %w", err)
	}

	if invitation.Status != "pending" {
		return errors.New("only pending invitations can be cancelled")
	}

	// Update status to cancelled
	invitation.Status = "cancelled"
	invitation.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	_, err = p.platformInvitationRepo.UpdateById(ctx, invitation.Id, invitation)
	if err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	return nil
}

func (p *platformDomain) IsTeamSlugAvailableForRequest(ctx context.Context, slug string) (bool, error) {
	// Check if there's a pending request with this slug
	existingRequest, err := p.teamApprovalRepo.FindOne(ctx, repos.Filter{
		"teamSlug": slug,
		"status":   entities.ApprovalStatusPending,
	})
	if err != nil && !errors.Is(err, repos.ErrNoDocuments) {
		return false, fmt.Errorf("failed to check team approval requests: %w", err)
	}
	
	// If there's a pending request, the slug is not available
	return existingRequest == nil, nil
}

func (p *platformDomain) AcceptPlatformInvitation(ctx context.Context, token string) error {
	invitation, err := p.platformInvitationRepo.FindOne(ctx, repos.Filter{
		"token": token,
	})
	if err != nil {
		if errors.Is(err, repos.ErrNoDocuments) {
			return errors.New("invalid invitation token")
		}
		return fmt.Errorf("failed to find invitation: %w", err)
	}

	if invitation.Status != "pending" {
		return errors.New("invitation is not pending")
	}

	if time.Now().After(invitation.ExpiresAt) {
		// Update status to expired
		invitation.Status = "expired"
		_, updateErr := p.platformInvitationRepo.UpdateById(ctx, invitation.Id, invitation)
		if updateErr != nil {
			p.logger.Error("failed to update expired invitation", "error", updateErr)
		}
		return errors.New("invitation has expired")
	}

	// Get or create user account
	userResp, err := p.authClient.EnsureUserByEmail(ctx, &authrpc.GetUserByEmailRequest{
		Email: invitation.Email,
	})
	if err != nil {
		return fmt.Errorf("failed to get user by email: %w", err)
	}

	// Create or update platform user in auth service
	createResp, err := p.authClient.CreateOrUpdatePlatformUser(ctx, &authrpc.CreateOrUpdatePlatformUserRequest{
		UserId: userResp.UserId,
		Email:  invitation.Email,
		Role:   string(invitation.Role),
	})
	if err != nil {
		return fmt.Errorf("failed to create platform user: %w", err)
	}
	
	if !createResp.Success {
		return errors.New("failed to create platform user")
	}

	// Update the invitation status
	invitation.Status = "accepted"
	now := time.Now()
	invitation.AcceptedAt = &now
	invitation.AcceptedBy = &userResp.UserId

	_, err = p.platformInvitationRepo.UpdateById(ctx, invitation.Id, invitation)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	p.logger.Info("platform invitation accepted", "email", invitation.Email, "role", invitation.Role, "userId", userResp.UserId)
	
	return nil
}