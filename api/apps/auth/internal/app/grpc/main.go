package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/apps/auth/internal/env"
	rpc_auth "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authGrpcServer struct {
	rpc_auth.UnimplementedAuthServer
	d                  domain.Domain
	env                *env.AuthEnv
	tokenExpiry        time.Duration
	refreshTokenExpiry time.Duration
	logger             *slog.Logger
}

type Claims struct {
	jwt.RegisteredClaims
	UserId string `json:"userId"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	UserId string `json:"userId"`
}

func (a *authGrpcServer) LoginWithOAuth(ctx context.Context, req *rpc_auth.LoginWithOAuthRequest) (*rpc_auth.LoginWithOAuthResponse, error) {
	user, err := a.d.LoginWithOAuth(ctx, req.Email, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not login with OAuth: %v", err)
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	
	token, refreshToken, err := a.generateTokens(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}
	
	return &rpc_auth.LoginWithOAuthResponse{
		UserId:       string(user.Id),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authGrpcServer) LoginWithSSO(ctx context.Context, req *rpc_auth.LoginWithSSORequest) (*rpc_auth.LoginWithSSOResponse, error) {
	user, err := a.d.LoginWithOAuth(ctx, req.Email, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not login with OAuth: %v", err)
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	
	token, refreshToken, err := a.generateTokens(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}
	
	return &rpc_auth.LoginWithSSOResponse{
		UserId:       string(user.Id),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authGrpcServer) GetUserDetails(ctx context.Context, req *rpc_auth.GetUserDetailsRequest) (*rpc_auth.GetUserDetailsResponse, error) {
	user, err := a.d.GetUserById(ctx, repos.ID(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get user details: %v", err)
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &rpc_auth.GetUserDetailsResponse{
		UserId:        string(user.Id),
		Name:          user.Name,
		Email:         user.Email,
		EmailVerified: user.Verified,
	}, nil
}

func (a *authGrpcServer) RequestResetPassword(ctx context.Context, req *rpc_auth.RequestResetPasswordRequest) (*rpc_auth.RequestResetPasswordResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	ok, err := a.d.RequestResetPassword(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not request password reset: %v", err)
	}
	if !ok {
		return nil, status.Error(codes.NotFound, "no account found for provided email")
	}

	return &rpc_auth.RequestResetPasswordResponse{
		Success: true,
	}, nil
}

// ResendEmailVerification implements v2.AuthV2Server.
func (a *authGrpcServer) ResendEmailVerification(ctx context.Context, req *rpc_auth.ResendEmailVerificationRequest) (*rpc_auth.ResendEmailVerificationResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	ok, err := a.d.ResendVerificationEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not resend verification email: %v", err)
	}
	if !ok {
		return nil, status.Error(codes.NotFound, "no account found for provided email")
	}

	return &rpc_auth.ResendEmailVerificationResponse{
		Success: true,
	}, nil
}

// VerifyEmail implements v2.AuthV2Server.
func (a *authGrpcServer) VerifyEmail(ctx context.Context, req *rpc_auth.VerifyEmailRequest) (*rpc_auth.VerifyEmailResponse, error) {
	if req.VerificationToken == "" {
		return nil, status.Error(codes.InvalidArgument, "verification token is required")
	}
	
	session, err := a.d.VerifyEmail(ctx, req.VerificationToken)
	if err != nil {
		a.logger.Error("email verification failed", "error", err, "token", req.VerificationToken)
		return nil, status.Errorf(codes.Internal, "failed to verify email: %v", err)
	}
	
	return &rpc_auth.VerifyEmailResponse{
		Success: true,
		UserId: string(session.UserId),
	}, nil
}

func (a *authGrpcServer) ResetPassword(ctx context.Context, resetReq *rpc_auth.ResetPasswordRequest) (*rpc_auth.ResetPasswordResponse, error) {
	done, err := a.d.ResetPassword(ctx, resetReq.ResetToken, resetReq.NewPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to reset password")
	}
	if !done {
		return nil, status.Error(codes.NotFound, "email not found")
	}
	return &rpc_auth.ResetPasswordResponse{
		Success: true,
	}, nil
}

func (a *authGrpcServer) Signup(ctx context.Context, signupReq *rpc_auth.SignupRequest) (*rpc_auth.SignupResponse, error) {
	if signupReq.Email == "" || signupReq.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	
	u, err := a.d.SignUp(ctx, signupReq.Name, signupReq.Email, signupReq.Password)
	if err != nil {
		errMsg := err.Error()
		// Check if the error is about user already existing
		if strings.Contains(errMsg, "already exists") {
			return nil, status.Error(codes.AlreadyExists, errMsg)
		}
		// Check for MongoDB duplicate key error
		if strings.Contains(errMsg, "duplicate key") || strings.Contains(errMsg, "E11000") {
			return nil, status.Error(codes.AlreadyExists, "An account with this email already exists")
		}
		// Check for validation errors
		if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "required") {
			return nil, status.Error(codes.InvalidArgument, errMsg)
		}
		// For other errors, log them but return a user-friendly message
		a.logger.Error("signup failed", "error", err, "email", signupReq.Email)
		return nil, status.Error(codes.Internal, "Unable to create account. Please try again later.")
	}
	
	
	token, refreshToken, err := a.generateTokens(u)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}
	
	return &rpc_auth.SignupResponse{
		UserId:       string(u.Id),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// JWT helper methods
func (a *authGrpcServer) generateTokens(user *entities.User) (token string, refreshToken string, err error) {
	// Generate access token
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "kloudlite-auth",
			Subject:   string(user.Id),
		},
		UserId: string(user.Id),
		Email:  user.Email,
		Name:   user.Name,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = accessToken.SignedString([]byte(a.env.JWTSecret))
	if err != nil {
		return "", "", errors.NewE(err)
	}

	// Generate refresh token
	refreshClaims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "kloudlite-auth",
			Subject:   string(user.Id),
		},
		UserId: string(user.Id),
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(a.env.JWTSecret))
	if err != nil {
		return "", "", errors.NewE(err)
	}

	return token, refreshToken, nil
}

func (a *authGrpcServer) validateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.env.JWTSecret), nil
	})

	if err != nil {
		return nil, errors.NewE(err)
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.Newf("invalid refresh token")
}

// RefreshToken implements the RefreshToken RPC
func (a *authGrpcServer) RefreshToken(ctx context.Context, req *rpc_auth.RefreshTokenRequest) (*rpc_auth.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	claims, err := a.validateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	// Get user from domain layer
	user, err := a.d.GetUserById(ctx, repos.ID(claims.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	token, refreshToken, err := a.generateTokens(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	return &rpc_auth.RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authGrpcServer) Login(ctx context.Context, loginRequest *rpc_auth.LoginRequest) (*rpc_auth.LoginResponse, error) {
	if loginRequest.Email == "" || loginRequest.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	
	user, err := a.d.Login(ctx, loginRequest.Email, loginRequest.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	
	token, refreshToken, err := a.generateTokens(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}
	
	return &rpc_auth.LoginResponse{
		UserId:       string(user.Id),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// Device Flow Methods
func (a *authGrpcServer) InitiateDeviceFlow(ctx context.Context, req *rpc_auth.InitiateDeviceFlowRequest) (*rpc_auth.InitiateDeviceFlowResponse, error) {
	if req.ClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "client ID is required")
	}

	deviceFlow, err := a.d.InitiateDeviceFlow(ctx, req.ClientId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to initiate device flow: %v", err)
	}

	// Calculate expires in seconds
	expiresIn := int32(time.Until(deviceFlow.ExpiresAt).Seconds())

	return &rpc_auth.InitiateDeviceFlowResponse{
		DeviceCode:             deviceFlow.DeviceCode,
		UserCode:               deviceFlow.UserCode,
		VerificationUri:        fmt.Sprintf("%s/device", a.env.WebUrl),
		VerificationUriComplete: fmt.Sprintf("%s/device?code=%s", a.env.WebUrl, deviceFlow.UserCode),
		ExpiresIn:              expiresIn,
		Interval:               5, // Poll every 5 seconds
	}, nil
}

func (a *authGrpcServer) PollDeviceToken(ctx context.Context, req *rpc_auth.PollDeviceTokenRequest) (*rpc_auth.PollDeviceTokenResponse, error) {
	if req.DeviceCode == "" || req.ClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "device code and client ID are required")
	}

	deviceFlow, err := a.d.PollDeviceToken(ctx, req.DeviceCode, req.ClientId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return &rpc_auth.PollDeviceTokenResponse{
				Error: "expired_token",
			}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to poll device token: %v", err)
	}

	// Check if expired
	if time.Now().After(deviceFlow.ExpiresAt) {
		return &rpc_auth.PollDeviceTokenResponse{
			Error: "expired_token",
		}, nil
	}

	// Check if authorized
	if !deviceFlow.Authorized {
		return &rpc_auth.PollDeviceTokenResponse{
			Error: "authorization_pending",
		}, nil
	}

	// Get user details
	user, err := a.d.GetUserById(ctx, repos.ID(deviceFlow.UserID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// Generate tokens
	token, refreshToken, err := a.generateTokens(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	return &rpc_auth.PollDeviceTokenResponse{
		Authorized:   true,
		UserId:       deviceFlow.UserID,
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authGrpcServer) VerifyDeviceCode(ctx context.Context, req *rpc_auth.VerifyDeviceCodeRequest) (*rpc_auth.VerifyDeviceCodeResponse, error) {
	if req.UserCode == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user code and user ID are required")
	}

	err := a.d.VerifyDeviceCode(ctx, req.UserCode, repos.ID(req.UserId))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "expired") {
			return &rpc_auth.VerifyDeviceCodeResponse{
				Success: false,
				Message: "Invalid or expired code",
			}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to verify device code: %v", err)
	}

	return &rpc_auth.VerifyDeviceCodeResponse{
		Success: true,
		Message: "Device successfully authorized",
	}, nil
}

// Platform user management methods
func (a *authGrpcServer) GetPlatformRole(ctx context.Context, req *rpc_auth.GetPlatformRoleRequest) (*rpc_auth.GetPlatformRoleResponse, error) {
	// Extract user ID from JWT token
	userId, err := a.getUserIdFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Get platform user
	platformUser, err := a.d.GetPlatformUser(ctx, userId)
	if err != nil {
		if errors.Is(err, repos.ErrNoDocuments) {
			// User has no platform role
			return &rpc_auth.GetPlatformRoleResponse{
				Role:              "",
				CanCreateTeams:    false,
				CanManagePlatform: false,
			}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to get platform role: %v", err)
	}

	role := string(platformUser.Role)
	canCreateTeams := role == "admin" || role == "super_admin"
	canManagePlatform := role == "admin" || role == "super_admin"

	return &rpc_auth.GetPlatformRoleResponse{
		Role:              role,
		CanCreateTeams:    canCreateTeams,
		CanManagePlatform: canManagePlatform,
	}, nil
}

func (a *authGrpcServer) ListPlatformUsers(ctx context.Context, req *rpc_auth.ListPlatformUsersRequest) (*rpc_auth.ListPlatformUsersResponse, error) {
	// Extract user ID from JWT token
	userId, err := a.getUserIdFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Check if user is platform admin
	platformUser, err := a.d.GetPlatformUser(ctx, userId)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	role := string(platformUser.Role)
	if role != "admin" && role != "super_admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only platform admins can list users")
	}

	// List platform users
	var roleFilter *entities.PlatformRole
	if req.Role != "" {
		r := entities.PlatformRole(req.Role)
		roleFilter = &r
	}

	platformUsers, err := a.d.ListPlatformUsers(ctx, roleFilter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list platform users: %v", err)
	}

	// Convert to response format
	users := make([]*rpc_auth.PlatformUser, 0, len(platformUsers))
	for _, pu := range platformUsers {
		// Get user details
		user, err := a.d.GetUserById(ctx, pu.UserId)
		if err != nil {
			continue
		}

		users = append(users, &rpc_auth.PlatformUser{
			UserId:    string(pu.UserId),
			Email:     user.Email,
			Role:      string(pu.Role),
			CreatedAt: pu.CreationTime.Format(time.RFC3339),
		})
	}

	return &rpc_auth.ListPlatformUsersResponse{
		Users: users,
	}, nil
}

func (a *authGrpcServer) UpdatePlatformUserRole(ctx context.Context, req *rpc_auth.UpdatePlatformUserRoleRequest) (*rpc_auth.UpdatePlatformUserRoleResponse, error) {
	// Extract user ID from JWT token
	userId, err := a.getUserIdFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Check if user is platform admin
	platformUser, err := a.d.GetPlatformUser(ctx, userId)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	role := string(platformUser.Role)
	if role != "admin" && role != "super_admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only platform admins can update user roles")
	}

	// Additional check: only super_admin can create other super_admins
	if req.Role == "super_admin" && role != "super_admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only super admins can create other super admins")
	}

	// Update platform user role
	err = a.d.CreateOrUpdatePlatformUser(ctx, &entities.PlatformUser{
		UserId: repos.ID(req.UserId),
		Role:   entities.PlatformRole(req.Role),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update platform user role: %v", err)
	}

	return &rpc_auth.UpdatePlatformUserRoleResponse{
		Success: true,
	}, nil
}

// Helper method to extract user ID from JWT context
func (a *authGrpcServer) getUserIdFromContext(ctx context.Context) (repos.ID, error) {
	// Extract metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.Newf("metadata not found")
	}

	// Get authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", errors.Newf("authorization header not found")
	}

	// Extract token from "Bearer <token>" format
	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.Newf("invalid authorization format")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate JWT token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Newf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.env.JWTSecret), nil
	})

	if err != nil {
		return "", errors.Newf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", errors.Newf("invalid token claims")
	}

	return repos.ID(claims.UserId), nil
}

// Notification methods
func (a *authGrpcServer) ListNotifications(ctx context.Context, req *rpc_auth.ListNotificationsRequest) (*rpc_auth.ListNotificationsResponse, error) {
	// Extract user ID from JWT token
	userId, err := a.getUserIdFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Get notifications
	notifications, totalCount, err := a.d.ListUserNotifications(ctx, userId, int(req.Limit), int(req.Offset), req.UnreadOnly, req.ActionRequiredOnly)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list notifications: %v", err)
	}

	// Convert to response format
	respNotifications := make([]*rpc_auth.Notification, 0, len(notifications))
	for _, notif := range notifications {
		// Determine unread status based on notification type
		isUnread := false
		if notif.Target.Type == entities.TargetTypeUser {
			isUnread = !notif.Read
		} else if notif.ActionRequired {
			isUnread = !notif.HasUserTakenAction(userId)
		}

		// Convert actions
		var actions []*rpc_auth.NotificationAction
		if len(notif.Actions) > 0 {
			actions = make([]*rpc_auth.NotificationAction, 0, len(notif.Actions))
			for _, action := range notif.Actions {
				actions = append(actions, &rpc_auth.NotificationAction{
					Id:       action.Id,
					Label:    action.Label,
					Style:    action.Style,
					Endpoint: action.Endpoint,
					Method:   action.Method,
					Data:     action.Data,
				})
			}
		}
		
		respNotifications = append(respNotifications, &rpc_auth.Notification{
			Id:             string(notif.Id),
			Title:          notif.Title,
			Description:    notif.Description,
			Type:           string(notif.Type),
			ActionRequired: notif.ActionRequired,
			Actions:        actions,
			ActionTaken:    notif.HasUserTakenAction(userId),
			CreatedAt:      notif.CreationTime.Format(time.RFC3339),
			Read:           !isUnread,
		})
	}

	return &rpc_auth.ListNotificationsResponse{
		Notifications: respNotifications,
		TotalCount:    int32(totalCount),
	}, nil
}

func (a *authGrpcServer) GetUnreadNotificationCount(ctx context.Context, req *rpc_auth.GetUnreadNotificationCountRequest) (*rpc_auth.GetUnreadNotificationCountResponse, error) {
	// Extract user ID from JWT token
	userId, err := a.getUserIdFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	count, err := a.d.GetUnreadNotificationCount(ctx, userId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get unread count: %v", err)
	}

	return &rpc_auth.GetUnreadNotificationCountResponse{
		Count: int32(count),
	}, nil
}

func (a *authGrpcServer) MarkNotificationAsRead(ctx context.Context, req *rpc_auth.MarkNotificationAsReadRequest) (*rpc_auth.MarkNotificationAsReadResponse, error) {
	// Extract user ID from JWT token
	userId, err := a.getUserIdFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	err = a.d.MarkNotificationAsRead(ctx, repos.ID(req.NotificationId), userId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark notification as read: %v", err)
	}

	return &rpc_auth.MarkNotificationAsReadResponse{
		Success: true,
	}, nil
}

func (a *authGrpcServer) MarkAllNotificationsAsRead(ctx context.Context, req *rpc_auth.MarkAllNotificationsAsReadRequest) (*rpc_auth.MarkAllNotificationsAsReadResponse, error) {
	// Extract user ID from JWT token
	userId, err := a.getUserIdFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	markedCount, err := a.d.MarkAllNotificationsAsRead(ctx, userId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark all notifications as read: %v", err)
	}

	return &rpc_auth.MarkAllNotificationsAsReadResponse{
		MarkedCount: int32(markedCount),
	}, nil
}

func (a *authGrpcServer) MarkNotificationActionTaken(ctx context.Context, req *rpc_auth.MarkNotificationActionTakenRequest) (*rpc_auth.MarkNotificationActionTakenResponse, error) {
	// Extract user ID from JWT token
	userId, err := a.getUserIdFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	err = a.d.MarkNotificationActionTaken(ctx, repos.ID(req.NotificationId), userId, req.ActionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark action taken: %v", err)
	}

	return &rpc_auth.MarkNotificationActionTakenResponse{
		Success: true,
	}, nil
}

func NewServer(d domain.Domain, env *env.AuthEnv, logger *slog.Logger) (rpc_auth.AuthServer, error) {
	tokenExpiry, err := time.ParseDuration(env.JWTTokenExpiry)
	if err != nil {
		return nil, errors.NewE(err)
	}

	refreshTokenExpiry, err := time.ParseDuration(env.JWTRefreshTokenExpiry)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &authGrpcServer{
		d:                  d,
		env:                env,
		tokenExpiry:        tokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		logger:             logger,
	}, nil
}
