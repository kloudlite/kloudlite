package grpc

import (
	"context"

	"github.com/kloudlite/api/apps/auth/internal/domain"
	rpc_auth "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/pkg/repos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authGrpcServer struct {
	rpc_auth.UnimplementedAuthServer
	d domain.Domain
	//sessionRepo kv.Repo[*common.AuthSession]
}

func (a *authGrpcServer) LoginWithOAuth(ctx context.Context, req *rpc_auth.LoginWithOAuthRequest) (*rpc_auth.LoginWithOAuthResponse, error) {
	user, err := a.d.LoginWithOAuth(ctx, req.Email, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not login with OAuth: %v", err)
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &rpc_auth.LoginWithOAuthResponse{
		UserId: string(user.Id),
	}, nil
}

func (a *authGrpcServer) LoginWithSSO(ctx context.Context, req *rpc_auth.LoginWithSSORequest) (*rpc_auth.LoginWithSSOResponse, error) {
	user, err := a.d.LoginWithSSO(ctx, req.Email, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not login with SSO: %v", err)
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &rpc_auth.LoginWithSSOResponse{
		UserId: string(user.Id),
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
	return &rpc_auth.VerifyEmailResponse{
		Success: true,
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
	u, err := a.d.SignUp(ctx, signupReq.Name, signupReq.Email, signupReq.Password)
	if signupReq.Email == "" || signupReq.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}
	return &rpc_auth.SignupResponse{
		UserId: string(u.Id),
	}, nil
}

func (a *authGrpcServer) Login(ctx context.Context, loginRequest *rpc_auth.LoginRequest) (*rpc_auth.LoginResponse, error) {
	user, err := a.d.Login(ctx, loginRequest.Email, loginRequest.Password)
	if loginRequest.Email == "" || loginRequest.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	return &rpc_auth.LoginResponse{
		UserId: string(user.Id),
	}, nil
}

func NewServer(d domain.Domain) rpc_auth.AuthServer {
	return &authGrpcServer{
		d: d,
	}
}
