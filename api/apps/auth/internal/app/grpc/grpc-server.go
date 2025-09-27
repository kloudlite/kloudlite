package grpc

import (
	"context"
	"time"

	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/api/pkg/functions"

	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/apps/auth/internal/app/email"
	authrpc "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/pkg/errors"
)

type authInternalGrpcServer struct {
	authrpc.UnimplementedAuthInternalServer
	d            domain.Domain
	emailService *email.EmailService
}

// GetAccessToken implements auth.AuthInternalServer.
func (a *authInternalGrpcServer) GetAccessToken(ctx context.Context, in *authrpc.GetAccessTokenRequest) (*authrpc.AccessTokenOut, error) {
	// TODO: Implement OAuth access token retrieval if needed
	return nil, errors.Newf("GetAccessToken not implemented")
}

// EnsureUserByEmail implements auth.AuthInternalServer.
func (a *authInternalGrpcServer) EnsureUserByEmail(ctx context.Context, in *authrpc.GetUserByEmailRequest) (*authrpc.GetUserByEmailOut, error) {
	user, err := a.d.GetUserByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, repos.ErrNoDocuments) {
		return nil, errors.NewE(err)
	}
	
	// If user doesn't exist, create a new one
	if user == nil {
		// Generate a temporary password - user will need to reset it
		tempPassword := functions.CleanerNanoidOrDie(16)
		user, err = a.d.SignUp(ctx, in.Email, in.Email, tempPassword)
		if err != nil {
			return nil, errors.NewEf(err, "failed to create user for email %q", in.Email)
		}
	}
	
	return &authrpc.GetUserByEmailOut{
		UserId: string(user.Id),
	}, nil
}

// GenerateMachineSession implements auth.AuthServer.
func (a *authInternalGrpcServer) GenerateMachineSession(ctx context.Context, in *authrpc.GenerateMachineSessionIn) (*authrpc.GenerateMachineSessionOut, error) {
	// TODO: Implement machine session with JWT tokens if needed
	return nil, errors.Newf("machine sessions not implemented with JWT")
}

// ClearMachineSessionByMachine implements auth.AuthServer.
func (a *authInternalGrpcServer) ClearMachineSessionByMachine(context.Context, *authrpc.ClearMachineSessionByMachineIn) (*authrpc.ClearMachineSessionByMachineOut, error) {
	panic("unimplemented")
}

// ClearMachineSessionByTeam implements auth.AuthServer.
func (a *authInternalGrpcServer) ClearMachineSessionByTeam(context.Context, *authrpc.ClearMachineSessionByTeamIn) (*authrpc.ClearMachineSessionByTeamOut, error) {
	panic("unimplemented")
}

// ClearMachineSessionByUser implements auth.AuthServer.
func (a *authInternalGrpcServer) ClearMachineSessionByUser(context.Context, *authrpc.ClearMachineSessionByUserIn) (*authrpc.ClearMachineSessionByUserOut, error) {
	panic("unimplemented")
}

func (a *authInternalGrpcServer) GetUser(ctx context.Context, in *authrpc.GetUserIn) (*authrpc.GetUserOut, error) {
	user, err := a.d.GetUserById(ctx, repos.ID(in.UserId))
	if err != nil {
		return nil, errors.NewE(err)
	}
	if user == nil {
		return nil, errors.Newf("could not find user with (id=%q)", in.UserId)
	}
	return &authrpc.GetUserOut{
		Id:    string(user.Id),
		Email: user.Email,
		Name:  user.Name,
	}, nil
}

// SendAccountInviteEmail implements auth.AuthInternalServer.
func (a *authInternalGrpcServer) SendAccountInviteEmail(ctx context.Context, request *authrpc.SendAccountInviteEmailRequest) (*authrpc.SendEmailResponse, error) {
	err := a.emailService.SendAccountInviteEmail(ctx, request.Email, request.Name, request.InvitedBy, request.AccountName, request.InviteLink)
	if err != nil {
		return &authrpc.SendEmailResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	return &authrpc.SendEmailResponse{
		Success: true,
	}, nil
}

// SendPlatformInviteEmail implements auth.AuthInternalServer.
func (a *authInternalGrpcServer) SendPlatformInviteEmail(ctx context.Context, request *authrpc.SendPlatformInviteEmailRequest) (*authrpc.SendEmailResponse, error) {
	err := a.emailService.SendPlatformInviteEmail(ctx, request.Email, request.Name, request.InvitedBy, request.Role, request.InviteLink)
	if err != nil {
		return &authrpc.SendEmailResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	return &authrpc.SendEmailResponse{
		Success: true,
	}, nil
}

// SendAlertEmail implements auth.AuthInternalServer.
func (a *authInternalGrpcServer) SendAlertEmail(ctx context.Context, request *authrpc.SendAlertEmailRequest) (*authrpc.SendEmailResponse, error) {
	// Convert protobuf map to Go map
	alertData := make(map[string]any)
	for k, v := range request.AlertData {
		alertData[k] = v
	}
	
	err := a.emailService.SendAlertEmail(ctx, request.Email, request.AlertTitle, request.AlertMessage, alertData)
	if err != nil {
		return &authrpc.SendEmailResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	return &authrpc.SendEmailResponse{
		Success: true,
	}, nil
}

// SendContactUsEmail implements auth.AuthInternalServer.
func (a *authInternalGrpcServer) SendContactUsEmail(ctx context.Context, request *authrpc.SendContactUsEmailRequest) (*authrpc.SendEmailResponse, error) {
	err := a.emailService.SendContactUsEmail(ctx, request.CustomerEmail, request.CustomerName, request.Subject, request.Message)
	if err != nil {
		return &authrpc.SendEmailResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	return &authrpc.SendEmailResponse{
		Success: true,
	}, nil
}

// Platform user methods
func (a *authInternalGrpcServer) GetPlatformUser(ctx context.Context, request *authrpc.GetPlatformUserRequest) (*authrpc.GetPlatformUserResponse, error) {
	platformUser, err := a.d.GetPlatformUser(ctx, repos.ID(request.UserId))
	if err != nil {
		// If no platform user found, return nil (not an error)
		if errors.Is(err, repos.ErrNoDocuments) {
			return &authrpc.GetPlatformUserResponse{
				PlatformUser: nil,
			}, nil
		}
		return nil, errors.NewE(err)
	}
	
	if platformUser == nil {
		return &authrpc.GetPlatformUserResponse{
			PlatformUser: nil,
		}, nil
	}
	
	// Get the user to fetch email
	user, err := a.d.GetUserById(ctx, platformUser.UserId)
	if err != nil {
		return nil, errors.NewE(err)
	}
	
	return &authrpc.GetPlatformUserResponse{
		PlatformUser: &authrpc.AuthPlatformUser{
			UserId:    string(platformUser.UserId),
			Email:     user.Email,
			Role:      string(platformUser.Role),
			CreatedAt: platformUser.CreationTime.Format(time.RFC3339),
		},
	}, nil
}

func (a *authInternalGrpcServer) CreateOrUpdatePlatformUser(ctx context.Context, req *authrpc.CreateOrUpdatePlatformUserRequest) (*authrpc.CreateOrUpdatePlatformUserResponse, error) {
	// Create or update platform user
	platformUser := &entities.PlatformUser{
		UserId:   repos.ID(req.UserId),
		Role:     entities.PlatformRole(req.Role),
	}
	
	err := a.d.CreateOrUpdatePlatformUser(ctx, platformUser)
	if err != nil {
		return nil, errors.NewE(err)
	}
	
	return &authrpc.CreateOrUpdatePlatformUserResponse{
		Success: true,
		UserId:  req.UserId,
	}, nil
}

func (a *authInternalGrpcServer) ListPlatformUsers(ctx context.Context, req *authrpc.InternalListPlatformUsersRequest) (*authrpc.InternalListPlatformUsersResponse, error) {
	var role *entities.PlatformRole
	if req.Role != "" {
		r := entities.PlatformRole(req.Role)
		role = &r
	}
	
	platformUsers, err := a.d.ListPlatformUsers(ctx, role)
	if err != nil {
		return nil, errors.NewE(err)
	}
	
	// Convert to response format
	users := make([]*authrpc.AuthPlatformUser, 0, len(platformUsers))
	for _, pu := range platformUsers {
		// Get user details
		user, err := a.d.GetUserById(ctx, pu.UserId)
		if err != nil {
			// Log error but continue
			continue
		}
		
		users = append(users, &authrpc.AuthPlatformUser{
			UserId:    string(pu.UserId),
			Email:     user.Email,
			Role:      string(pu.Role),
			CreatedAt: pu.CreationTime.Format(time.RFC3339),
		})
	}
	
	return &authrpc.InternalListPlatformUsersResponse{
		Users: users,
	}, nil
}

// CreateNotification implements auth.AuthInternalServer
func (a *authInternalGrpcServer) CreateNotification(ctx context.Context, req *authrpc.CreateNotificationRequest) (*authrpc.CreateNotificationResponse, error) {
	// Build notification entity
	notification := &entities.Notification{
		Title:          req.Title,
		Description:    req.Description,
		Type:           entities.NotificationType(req.Type),
		ActionRequired: req.ActionRequired,
		DedupeKey:      req.DedupeKey,
	}
	
	// Convert actions from proto to entity
	if len(req.Actions) > 0 {
		notification.Actions = make([]entities.NotificationAction, 0, len(req.Actions))
		for _, action := range req.Actions {
			notification.Actions = append(notification.Actions, entities.NotificationAction{
				Id:       action.Id,
				Label:    action.Label,
				Style:    action.Style,
				Endpoint: action.Endpoint,
				Method:   action.Method,
				Data:     action.Data,
			})
		}
	}

	// Set target based on type
	target := entities.NotificationTarget{
		Type: entities.TargetType(req.Target.Type),
	}

	switch req.Target.Type {
	case string(entities.TargetTypeUser):
		if req.Target.UserId != "" {
			userId := repos.ID(req.Target.UserId)
			target.UserId = &userId
		}
	case string(entities.TargetTypeTeamRole):
		if req.Target.TeamId != "" {
			teamId := repos.ID(req.Target.TeamId)
			target.TeamId = &teamId
		}
		if req.Target.MinTeamRole != "" {
			target.MinTeamRole = &req.Target.MinTeamRole
		}
	case string(entities.TargetTypePlatformRole):
		if req.Target.MinPlatformRole != "" {
			target.MinPlatformRole = &req.Target.MinPlatformRole
		}
	}

	notification.Target = target

	// Create notification
	created, err := a.d.CreateNotification(ctx, notification)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &authrpc.CreateNotificationResponse{
		NotificationId: string(created.Id),
		Success:        true,
	}, nil
}

func NewInternalServer(d domain.Domain, emailService *email.EmailService) authrpc.AuthInternalServer {
	return &authInternalGrpcServer{
		d:            d,
		emailService: emailService,
	}
}
