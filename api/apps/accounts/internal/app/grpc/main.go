package grpc

import (
	"context"
	"time"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"github.com/kloudlite/api/pkg/repos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/api/apps/accounts/internal/app/jwt"
	"github.com/kloudlite/api/apps/accounts/internal/domain"
	"github.com/kloudlite/api/apps/accounts/internal/entities"
	"github.com/kloudlite/api/common"
)

type accountsGrpcServer struct {
	accounts.UnimplementedAccountsServer
	d             domain.Domain
	jwtInterceptor *jwt.JWTInterceptor
}

// validateJWT validates the JWT token and returns the user context
func (a *accountsGrpcServer) validateJWT(ctx context.Context) (*domain.UserContext, error) {
	// Create a fake unary server info
	info := &grpc.UnaryServerInfo{
		FullMethod: "/Accounts/ValidateJWT",
	}
	
	// Use the JWT interceptor to validate
	var userCtx *domain.UserContext
	_, err := a.jwtInterceptor.UnaryServerInterceptor()(
		ctx,
		nil,
		info,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			// After JWT validation, extract user context
			var extractErr error
			userCtx, extractErr = getUserContext(ctx)
			return nil, extractErr
		},
	)
	
	if err != nil {
		return nil, err
	}
	
	return userCtx, nil
}

func (a *accountsGrpcServer) CheckTeamSlugAvailability(ctx context.Context, request *accounts.CheckTeamSlugAvailabilityRequest) (*accounts.CheckTeamSlugAvailabilityResponse, error) {
	availability, err := a.d.CheckTeamSlugAvailability(ctx, request.Slug)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check team slug availability: %v", err)
	}
	if availability == nil {
		return nil, status.Errorf(codes.Internal, "check team slug availability returned nil")
	}
	return &accounts.CheckTeamSlugAvailabilityResponse{
		Result:         availability.Result,
		SuggestedSlugs: availability.SuggestedNames,
	}, nil
}

func (a *accountsGrpcServer) GenerateTeamSlugSuggestions(ctx context.Context, request *accounts.GenerateTeamSlugSuggestionsRequest) (*accounts.GenerateTeamSlugSuggestionsResponse, error) {
	suggestions := a.d.GenerateTeamSlugSuggestions(ctx, request.DisplayName)
	return &accounts.GenerateTeamSlugSuggestionsResponse{
		Suggestions: suggestions,
	}, nil
}

func getUserContext(ctx context.Context) (*domain.UserContext, error) {
	userId, userEmail, userName, err := jwt.ExtractUserContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to extract user context: %v", err)
	}
	return &domain.UserContext{
		Context:   ctx,
		UserId:    repos.ID(userId),
		UserName:  userName,
		UserEmail: userEmail,
	}, nil
}

func (a *accountsGrpcServer) CreateTeam(ctx context.Context, req *accounts.CreateTeamRequest) (*accounts.CreateTeamResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}
	newTeam := entities.Team{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Slug,
		},
		ResourceMetadata: common.ResourceMetadata{
			DisplayName: req.DisplayName,
		},
		Slug:        req.Slug,
		Description: req.Description,
		Region:      req.Region,
	}
	team, err := a.d.CreateTeam(*userContext, newTeam)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create team: %v", err)
	}
	if team == nil {
		return nil, status.Errorf(codes.Internal, "team creation returned nil")
	}
	return &accounts.CreateTeamResponse{
		TeamId: string(team.Id),
	}, nil
}

func (a *accountsGrpcServer) DeleteTeam(context.Context, *accounts.DeleteTeamRequest) (*accounts.DeleteTeamResponse, error) {
	panic("unimplemented")
}

// DisableTeam implements accounts.AccountsServer.
func (a *accountsGrpcServer) DisableTeam(context.Context, *accounts.DisableTeamRequest) (*accounts.DisableTeamResponse, error) {
	panic("unimplemented")
}

// EnableTeam implements accounts.AccountsServer.
func (a *accountsGrpcServer) EnableTeam(context.Context, *accounts.EnableTeamRequest) (*accounts.EnableTeamResponse, error) {
	panic("unimplemented")
}

// GetTeamDetails implements accounts.AccountsServer.
func (a *accountsGrpcServer) GetTeamDetails(ctx context.Context, req *accounts.GetTeamDetailsRequest) (*accounts.GetTeamDetailsResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	team, err := a.d.GetTeam(*userContext, repos.ID(req.TeamId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get team: %v", err)
	}

	// We can skip role in GetTeamDetails since the user already has access

	return &accounts.GetTeamDetailsResponse{
		TeamId:      string(team.Id),
		Slug:        team.Slug,
		DisplayName: team.DisplayName,
		OwnerId:     string(team.OwnerId),
		Status:      "active",
		Region:      team.Region,
	}, nil
}

// ListTeams implements accounts.AccountsServer.
func (a *accountsGrpcServer) ListTeams(ctx context.Context, req *accounts.ListTeamsRequest) (*accounts.ListTeamsResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	teams, err := a.d.ListTeams(*userContext)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list teams: %v", err)
	}

	teamDetails := make([]*accounts.TeamDetails, len(teams))
	for i, team := range teams {
		role, err := a.d.GetUserRoleInTeam(ctx, userContext.UserId, team.Id)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get user role: %v", err)
		}

		status := "active"
		if team.IsActive != nil && !*team.IsActive {
			status = "inactive"
		}

		teamDetails[i] = &accounts.TeamDetails{
			TeamId:      string(team.Id),
			Slug:        team.Slug,
			DisplayName: team.DisplayName,
			OwnerId:     string(team.OwnerId),
			Status:      status,
			Region:      team.Region,
			Role:        string(role),
		}
	}

	return &accounts.ListTeamsResponse{
		Teams: teamDetails,
	}, nil
}

// SearchTeams implements accounts.AccountsServer.
func (a *accountsGrpcServer) SearchTeams(ctx context.Context, req *accounts.SearchTeamsRequest) (*accounts.SearchTeamsResponse, error) {
	teams, totalCount, err := a.d.SearchTeams(ctx, req.Query, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search teams: %v", err)
	}

	teamDetails := make([]*accounts.TeamDetails, len(teams))
	for i, team := range teams {
		status := "active"
		if team.IsActive != nil && !*team.IsActive {
			status = "inactive"
		}

		teamDetails[i] = &accounts.TeamDetails{
			TeamId:      string(team.Id),
			Slug:        team.Slug,
			DisplayName: team.DisplayName,
			OwnerId:     string(team.OwnerId),
			Status:      status,
			Region:      team.Region,
		}
	}

	return &accounts.SearchTeamsResponse{
		Teams:      teamDetails,
		TotalCount: int32(totalCount),
	}, nil
}

// GetUserTeams implements accounts.AccountsServer.
func (a *accountsGrpcServer) GetUserTeams(ctx context.Context, req *accounts.GetUserTeamsRequest) (*accounts.GetUserTeamsResponse, error) {
	userContext := domain.UserContext{
		Context: ctx,
		UserId: repos.ID(req.UserId),
	}

	teams, err := a.d.ListTeams(userContext)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list user teams: %v", err)
	}

	teamDetails := make([]*accounts.TeamDetails, len(teams))
	for i, team := range teams {
		role, err := a.d.GetUserRoleInTeam(ctx, userContext.UserId, team.Id)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get user role: %v", err)
		}

		status := "active"
		if team.IsActive != nil && !*team.IsActive {
			status = "inactive"
		}

		teamDetails[i] = &accounts.TeamDetails{
			TeamId:      string(team.Id),
			Slug:        team.Slug,
			DisplayName: team.DisplayName,
			OwnerId:     string(team.OwnerId),
			Status:      status,
			Region:      team.Region,
			Role:        string(role),
		}
	}

	return &accounts.GetUserTeamsResponse{
		Teams: teamDetails,
	}, nil
}

// GetTeamMembers implements accounts.AccountsServer.
func (a *accountsGrpcServer) GetTeamMembers(ctx context.Context, req *accounts.GetTeamMembersRequest) (*accounts.GetTeamMembersResponse, error) {
	// TODO: Implement team members listing
	return nil, status.Errorf(codes.Unimplemented, "GetTeamMembers not implemented")
}

// InviteTeamMember implements accounts.AccountsServer.
func (a *accountsGrpcServer) InviteTeamMember(ctx context.Context, req *accounts.InviteTeamMemberRequest) (*accounts.InviteTeamMemberResponse, error) {
	// TODO: Implement team member invitation
	return nil, status.Errorf(codes.Unimplemented, "InviteTeamMember not implemented")
}

// RemoveTeamMember implements accounts.AccountsServer.
func (a *accountsGrpcServer) RemoveTeamMember(ctx context.Context, req *accounts.RemoveTeamMemberRequest) (*accounts.RemoveTeamMemberResponse, error) {
	// TODO: Implement team member removal
	return nil, status.Errorf(codes.Unimplemented, "RemoveTeamMember not implemented")
}

// UpdateTeamMemberRole implements accounts.AccountsServer.
func (a *accountsGrpcServer) UpdateTeamMemberRole(ctx context.Context, req *accounts.UpdateTeamMemberRoleRequest) (*accounts.UpdateTeamMemberRoleResponse, error) {
	// TODO: Implement team member role update
	return nil, status.Errorf(codes.Unimplemented, "UpdateTeamMemberRole not implemented")
}

// Platform management implementations

func (a *accountsGrpcServer) RequestTeamCreation(ctx context.Context, req *accounts.RequestTeamCreationRequest) (*accounts.RequestTeamCreationResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	request := entities.TeamApprovalRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Slug,
		},
		ResourceMetadata: common.ResourceMetadata{
			DisplayName: req.DisplayName,
		},
		TeamSlug:        req.Slug,
		TeamDescription: req.Description,
		TeamRegion:      req.Region,
	}

	createdRequest, err := a.d.RequestTeamCreation(*userContext, request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create team request: %v", err)
	}

	return &accounts.RequestTeamCreationResponse{
		RequestId: string(createdRequest.Id),
		Status:    string(createdRequest.Status),
	}, nil
}

func (a *accountsGrpcServer) ListTeamRequests(ctx context.Context, req *accounts.ListTeamRequestsRequest) (*accounts.ListTeamRequestsResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	var statusFilter *entities.ApprovalStatus
	if req.Status != "" {
		status := entities.ApprovalStatus(req.Status)
		statusFilter = &status
	}

	requests, err := a.d.ListTeamApprovalRequests(*userContext, statusFilter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list team requests: %v", err)
	}

	teamRequests := make([]*accounts.TeamRequest, len(requests))
	for i, req := range requests {
		tr := &accounts.TeamRequest{
			RequestId:       string(req.Id),
			Slug:            req.TeamSlug,
			DisplayName:     req.DisplayName,
			Description:     req.TeamDescription,
			Region:          req.TeamRegion,
			Status:          string(req.Status),
			RequestedBy:     string(req.RequestedBy),
			RequestedByEmail: req.RequestedByEmail,
			RequestedAt:     req.RequestedAt.Format(time.RFC3339),
		}
		
		if req.ReviewedBy != nil {
			tr.ReviewedBy = string(*req.ReviewedBy)
		}
		if req.ReviewedByEmail != nil {
			tr.ReviewedByEmail = *req.ReviewedByEmail
		}
		if req.ReviewedAt != nil {
			tr.ReviewedAt = req.ReviewedAt.Format(time.RFC3339)
		}
		if req.RejectionReason != nil {
			tr.RejectionReason = *req.RejectionReason
		}
		
		teamRequests[i] = tr
	}

	return &accounts.ListTeamRequestsResponse{
		Requests: teamRequests,
	}, nil
}

func (a *accountsGrpcServer) GetTeamRequest(ctx context.Context, req *accounts.GetTeamRequestRequest) (*accounts.GetTeamRequestResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	request, err := a.d.GetTeamApprovalRequest(*userContext, repos.ID(req.RequestId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get team request: %v", err)
	}

	tr := &accounts.TeamRequest{
		RequestId:       string(request.Id),
		Slug:            request.TeamSlug,
		DisplayName:     request.DisplayName,
		Description:     request.TeamDescription,
		Region:          request.TeamRegion,
		Status:          string(request.Status),
		RequestedBy:     string(request.RequestedBy),
		RequestedByEmail: request.RequestedByEmail,
		RequestedAt:     request.RequestedAt.Format(time.RFC3339),
	}
	
	if request.ReviewedBy != nil {
		tr.ReviewedBy = string(*request.ReviewedBy)
	}
	if request.ReviewedByEmail != nil {
		tr.ReviewedByEmail = *request.ReviewedByEmail
	}
	if request.ReviewedAt != nil {
		tr.ReviewedAt = request.ReviewedAt.Format(time.RFC3339)
	}
	if request.RejectionReason != nil {
		tr.RejectionReason = *request.RejectionReason
	}

	return &accounts.GetTeamRequestResponse{
		Request: tr,
	}, nil
}

func (a *accountsGrpcServer) ApproveTeamRequest(ctx context.Context, req *accounts.ApproveTeamRequestRequest) (*accounts.ApproveTeamRequestResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	team, err := a.d.ApproveTeamRequest(*userContext, repos.ID(req.RequestId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to approve team request: %v", err)
	}

	return &accounts.ApproveTeamRequestResponse{
		TeamId: string(team.Id),
	}, nil
}

func (a *accountsGrpcServer) RejectTeamRequest(ctx context.Context, req *accounts.RejectTeamRequestRequest) (*accounts.RejectTeamRequestResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	_, err = a.d.RejectTeamRequest(*userContext, repos.ID(req.RequestId), req.Reason)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reject team request: %v", err)
	}

	return &accounts.RejectTeamRequestResponse{
		Success: true,
	}, nil
}

// Platform settings implementations

func (a *accountsGrpcServer) GetPlatformSettings(ctx context.Context, req *accounts.GetPlatformSettingsRequest) (*accounts.GetPlatformSettingsResponse, error) {
	// GetPlatformSettings doesn't require authentication as it's needed for login page
	settings, err := a.d.GetPlatformSettings(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get platform settings: %v", err)
	}

	// Check if this is an internal server request (has specific header)
	md, ok := metadata.FromIncomingContext(ctx)
	includeSecrets := ok && len(md.Get("x-internal-request")) > 0

	return &accounts.GetPlatformSettingsResponse{
		Settings: &accounts.PlatformSettings{
			PlatformOwnerEmail: settings.PlatformOwnerEmail,
			SupportEmail:       settings.SupportEmail,
			AllowSignup:        settings.AllowSignup,
			OauthProviders: &accounts.OAuthProviderSettings{
				Google: &accounts.OAuthProvider{
					Enabled:      settings.OAuthProviders.Google.Enabled,
					ClientId:     settings.OAuthProviders.Google.ClientId,
					ClientSecret: func() string {
						if includeSecrets {
							return settings.OAuthProviders.Google.ClientSecret
						}
						return ""
					}(),
				},
				Github: &accounts.OAuthProvider{
					Enabled:      settings.OAuthProviders.GitHub.Enabled,
					ClientId:     settings.OAuthProviders.GitHub.ClientId,
					ClientSecret: func() string {
						if includeSecrets {
							return settings.OAuthProviders.GitHub.ClientSecret
						}
						return ""
					}(),
				},
				Microsoft: &accounts.MicrosoftOAuthProvider{
					Enabled:      settings.OAuthProviders.Microsoft.Enabled,
					ClientId:     settings.OAuthProviders.Microsoft.ClientId,
					ClientSecret: func() string {
						if includeSecrets {
							return settings.OAuthProviders.Microsoft.ClientSecret
						}
						return ""
					}(),
					TenantId:     settings.OAuthProviders.Microsoft.TenantId,
				},
			},
			TeamSettings: &accounts.TeamSettings{
				RequireApproval:      settings.TeamSettings.RequireApproval,
				AutoApproveFirstTeam: settings.TeamSettings.AutoApproveFirstTeam,
				MaxTeamsPerUser:      int32(settings.TeamSettings.MaxTeamsPerUser),
			},
			Features: &accounts.PlatformFeatures{
				EnableDeviceFlow: settings.Features.EnableDeviceFlow,
				EnableCLI:        settings.Features.EnableCLI,
				EnableAPI:        settings.Features.EnableAPI,
			},
			CloudProvider: &accounts.CloudProviderConfig{
				Provider: settings.CloudProvider.Provider,
				Aws: &accounts.AWSConfig{
					AccessKeyId: settings.CloudProvider.AWS.AccessKeyId,
					SecretAccessKey: func() string {
						if includeSecrets {
							return settings.CloudProvider.AWS.SecretAccessKey
						}
						return ""
					}(),
					Region: settings.CloudProvider.AWS.Region,
				},
				Gcp: &accounts.GCPConfig{
					ProjectId: settings.CloudProvider.GCP.ProjectId,
					ServiceAccountKey: func() string {
						if includeSecrets {
							return settings.CloudProvider.GCP.ServiceAccountKey
						}
						return ""
					}(),
				},
				Azure: &accounts.AzureConfig{
					SubscriptionId: settings.CloudProvider.Azure.SubscriptionId,
					TenantId:       settings.CloudProvider.Azure.TenantId,
					ClientId:       settings.CloudProvider.Azure.ClientId,
					ClientSecret: func() string {
						if includeSecrets {
							return settings.CloudProvider.Azure.ClientSecret
						}
						return ""
					}(),
				},
				Digitalocean: &accounts.DigitalOceanConfig{
					Token: func() string {
						if includeSecrets {
							return settings.CloudProvider.DigitalOcean.Token
						}
						return ""
					}(),
				},
			},
		},
	}, nil
}

func (a *accountsGrpcServer) UpdatePlatformSettings(ctx context.Context, req *accounts.UpdatePlatformSettingsRequest) (*accounts.UpdatePlatformSettingsResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	settings := entities.PlatformSettings{
		PlatformOwnerEmail: req.Settings.PlatformOwnerEmail,
		SupportEmail:       req.Settings.SupportEmail,
		AllowSignup:        req.Settings.AllowSignup,
		OAuthProviders: struct {
			Google    entities.OAuthProvider `json:"google"`
			GitHub    entities.OAuthProvider `json:"github"`
			Microsoft entities.OAuthProvider `json:"microsoft"`
		}{
			Google: entities.OAuthProvider{
				Enabled:      req.Settings.OauthProviders.Google.Enabled,
				ClientId:     req.Settings.OauthProviders.Google.ClientId,
				ClientSecret: req.Settings.OauthProviders.Google.ClientSecret,
			},
			GitHub: entities.OAuthProvider{
				Enabled:      req.Settings.OauthProviders.Github.Enabled,
				ClientId:     req.Settings.OauthProviders.Github.ClientId,
				ClientSecret: req.Settings.OauthProviders.Github.ClientSecret,
			},
			Microsoft: entities.OAuthProvider{
				Enabled:      req.Settings.OauthProviders.Microsoft.Enabled,
				ClientId:     req.Settings.OauthProviders.Microsoft.ClientId,
				ClientSecret: req.Settings.OauthProviders.Microsoft.ClientSecret,
			},
		},
		TeamSettings: struct {
			RequireApproval      bool `json:"requireApproval"`
			AutoApproveFirstTeam bool `json:"autoApproveFirstTeam"`
			MaxTeamsPerUser      int  `json:"maxTeamsPerUser"`
		}{
			RequireApproval:      req.Settings.TeamSettings.RequireApproval,
			AutoApproveFirstTeam: req.Settings.TeamSettings.AutoApproveFirstTeam,
			MaxTeamsPerUser:      int(req.Settings.TeamSettings.MaxTeamsPerUser),
		},
		Features: struct {
			EnableDeviceFlow bool `json:"enableDeviceFlow"`
			EnableCLI        bool `json:"enableCLI"`
			EnableAPI        bool `json:"enableAPI"`
		}{
			EnableDeviceFlow: req.Settings.Features.EnableDeviceFlow,
			EnableCLI:        req.Settings.Features.EnableCLI,
			EnableAPI:        req.Settings.Features.EnableAPI,
		},
		CloudProvider: struct {
			Provider string `json:"provider,omitempty"`
			AWS struct {
				AccessKeyId     string `json:"accessKeyId,omitempty"`
				SecretAccessKey string `json:"secretAccessKey,omitempty"`
				Region          string `json:"region,omitempty"`
			} `json:"aws,omitempty"`
			GCP struct {
				ProjectId         string `json:"projectId,omitempty"`
				ServiceAccountKey string `json:"serviceAccountKey,omitempty"`
			} `json:"gcp,omitempty"`
			Azure struct {
				SubscriptionId string `json:"subscriptionId,omitempty"`
				TenantId       string `json:"tenantId,omitempty"`
				ClientId       string `json:"clientId,omitempty"`
				ClientSecret   string `json:"clientSecret,omitempty"`
			} `json:"azure,omitempty"`
			DigitalOcean struct {
				Token string `json:"token,omitempty"`
			} `json:"digitalocean,omitempty"`
		}{
			Provider: func() string { if req.Settings.CloudProvider != nil { return req.Settings.CloudProvider.Provider } else { return "" } }(),
			AWS: struct {
				AccessKeyId     string `json:"accessKeyId,omitempty"`
				SecretAccessKey string `json:"secretAccessKey,omitempty"`
				Region          string `json:"region,omitempty"`
			}{
				AccessKeyId:     func() string { if req.Settings.CloudProvider != nil && req.Settings.CloudProvider.Aws != nil { return req.Settings.CloudProvider.Aws.AccessKeyId } else { return "" } }(),
				SecretAccessKey: func() string { if req.Settings.CloudProvider != nil && req.Settings.CloudProvider.Aws != nil { return req.Settings.CloudProvider.Aws.SecretAccessKey } else { return "" } }(),
				Region:          func() string { if req.Settings.CloudProvider != nil && req.Settings.CloudProvider.Aws != nil { return req.Settings.CloudProvider.Aws.Region } else { return "" } }(),
			},
			GCP: struct {
				ProjectId         string `json:"projectId,omitempty"`
				ServiceAccountKey string `json:"serviceAccountKey,omitempty"`
			}{
				ProjectId:         func() string { if req.Settings.CloudProvider != nil && req.Settings.CloudProvider.Gcp != nil { return req.Settings.CloudProvider.Gcp.ProjectId } else { return "" } }(),
				ServiceAccountKey: func() string { if req.Settings.CloudProvider != nil && req.Settings.CloudProvider.Gcp != nil { return req.Settings.CloudProvider.Gcp.ServiceAccountKey } else { return "" } }(),
			},
			Azure: struct {
				SubscriptionId string `json:"subscriptionId,omitempty"`
				TenantId       string `json:"tenantId,omitempty"`
				ClientId       string `json:"clientId,omitempty"`
				ClientSecret   string `json:"clientSecret,omitempty"`
			}{
				SubscriptionId: func() string { if req.Settings.CloudProvider != nil && req.Settings.CloudProvider.Azure != nil { return req.Settings.CloudProvider.Azure.SubscriptionId } else { return "" } }(),
				TenantId:       func() string { if req.Settings.CloudProvider != nil && req.Settings.CloudProvider.Azure != nil { return req.Settings.CloudProvider.Azure.TenantId } else { return "" } }(),
				ClientId:       func() string { if req.Settings.CloudProvider != nil && req.Settings.CloudProvider.Azure != nil { return req.Settings.CloudProvider.Azure.ClientId } else { return "" } }(),
				ClientSecret:   func() string { if req.Settings.CloudProvider != nil && req.Settings.CloudProvider.Azure != nil { return req.Settings.CloudProvider.Azure.ClientSecret } else { return "" } }(),
			},
			DigitalOcean: struct {
				Token string `json:"token,omitempty"`
			}{
				Token: func() string { if req.Settings.CloudProvider != nil && req.Settings.CloudProvider.Digitalocean != nil { return req.Settings.CloudProvider.Digitalocean.Token } else { return "" } }(),
			},
		},
	}

	_, err = a.d.UpdatePlatformSettings(*userContext, settings)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update platform settings: %v", err)
	}

	return &accounts.UpdatePlatformSettingsResponse{
		Success: true,
	}, nil
}

// Platform invitation implementations

func (a *accountsGrpcServer) InvitePlatformUser(ctx context.Context, req *accounts.InvitePlatformUserRequest) (*accounts.InvitePlatformUserResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	invitation, err := a.d.InvitePlatformUser(*userContext, req.Email, req.Role)
	if err != nil {
		return &accounts.InvitePlatformUserResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &accounts.InvitePlatformUserResponse{
		Success:      true,
		InvitationId: string(invitation.Id),
	}, nil
}

func (a *accountsGrpcServer) ListPlatformInvitations(ctx context.Context, req *accounts.ListPlatformInvitationsRequest) (*accounts.ListPlatformInvitationsResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	var statusFilter *string
	if req.Status != "" {
		statusFilter = &req.Status
	}

	invitations, err := a.d.ListPlatformInvitations(*userContext, statusFilter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list platform invitations: %v", err)
	}

	grpcInvitations := make([]*accounts.PlatformInvitation, len(invitations))
	for i, inv := range invitations {
		grpcInv := &accounts.PlatformInvitation{
			Id:             string(inv.Id),
			Email:          inv.Email,
			Role:           inv.Role,
			InvitedBy:      inv.InvitedBy,
			InvitedByEmail: inv.InvitedByEmail,
			Status:         inv.Status,
			CreatedAt:      inv.CreationTime.Format(time.RFC3339),
			ExpiresAt:      inv.ExpiresAt.Format(time.RFC3339),
		}
		
		if inv.AcceptedAt != nil {
			grpcInv.AcceptedAt = inv.AcceptedAt.Format(time.RFC3339)
		}
		
		grpcInvitations[i] = grpcInv
	}

	return &accounts.ListPlatformInvitationsResponse{
		Invitations: grpcInvitations,
	}, nil
}

func (a *accountsGrpcServer) ResendPlatformInvitation(ctx context.Context, req *accounts.ResendPlatformInvitationRequest) (*accounts.ResendPlatformInvitationResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	err = a.d.ResendPlatformInvitation(*userContext, repos.ID(req.InvitationId))
	if err != nil {
		return &accounts.ResendPlatformInvitationResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &accounts.ResendPlatformInvitationResponse{
		Success: true,
	}, nil
}

func (a *accountsGrpcServer) CancelPlatformInvitation(ctx context.Context, req *accounts.CancelPlatformInvitationRequest) (*accounts.CancelPlatformInvitationResponse, error) {
	userContext, err := a.validateJWT(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate JWT: %v", err)
	}

	err = a.d.CancelPlatformInvitation(*userContext, repos.ID(req.InvitationId))
	if err != nil {
		return &accounts.CancelPlatformInvitationResponse{
			Success: false,
		}, status.Errorf(codes.Internal, "failed to cancel invitation: %v", err)
	}

	return &accounts.CancelPlatformInvitationResponse{
		Success: true,
	}, nil
}

func (a *accountsGrpcServer) AcceptPlatformInvitation(ctx context.Context, req *accounts.AcceptPlatformInvitationRequest) (*accounts.AcceptPlatformInvitationResponse, error) {
	err := a.d.AcceptPlatformInvitation(ctx, req.Token)
	if err != nil {
		return &accounts.AcceptPlatformInvitationResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &accounts.AcceptPlatformInvitationResponse{
		Success: true,
	}, nil
}

func NewServer(d domain.Domain, jwtInterceptor *jwt.JWTInterceptor) accounts.AccountsServer {
	return &accountsGrpcServer{
		d: d,
		jwtInterceptor: jwtInterceptor,
	}
}
