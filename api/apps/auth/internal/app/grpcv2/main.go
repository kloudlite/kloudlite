package grpcv2

import (
	"context"

	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/common"
	authV2 "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth/v2"
	"github.com/kloudlite/api/pkg/kv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authGrpcServer struct {
	authV2.UnimplementedAuthV2Server
	d           domain.Domain
	sessionRepo kv.Repo[*common.AuthSession]
}

// RequestResetPassword implements v2.AuthV2Server.
func (a *authGrpcServer) RequestResetPassword(ctx context.Context, req *authV2.RequestResetPasswordRequest) (*authV2.RequestResetPasswordResponse, error) {
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

	return &authV2.RequestResetPasswordResponse{
		Success: true,
	}, nil
}

// ResendEmailVerification implements v2.AuthV2Server.
func (a *authGrpcServer) ResendEmailVerification(ctx context.Context, req *authV2.ResendEmailVerificationRequest) (*authV2.ResendEmailVerificationResponse, error) {
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

	return &authV2.ResendEmailVerificationResponse{
		Success: true,
	}, nil
}

// VerifyEmail implements v2.AuthV2Server.
func (a *authGrpcServer) VerifyEmail(ctx context.Context, req *authV2.VerifyEmailRequest) (*authV2.VerifyEmailResponse, error) {
	if req.VerificationToken == "" {
		return nil, status.Error(codes.InvalidArgument, "verification token is required")
	}

	session, err := a.d.VerifyEmail(ctx, req.VerificationToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not verify email: %v", err)
	}

	if err := a.sessionRepo.Set(ctx, string(session.Id), session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store session: %v", err)
	}

	return &authV2.VerifyEmailResponse{
		Success: true,
	}, nil
}

func (a *authGrpcServer) ResetPassword(ctx context.Context, resetReq *authV2.ResetPasswordRequest) (*authV2.ResetPasswordResponse, error) {
	done, err := a.d.ResetPassword(ctx, resetReq.ResetToken, resetReq.NewPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to reset password")
	}
	if !done {
		return nil, status.Error(codes.NotFound, "email not found")
	}
	return &authV2.ResetPasswordResponse{
		Success: true,
	}, nil
}

func (a *authGrpcServer) Signup(ctx context.Context, signupReq *authV2.SignupRequest) (*authV2.SignupResponse, error) {
	u, err := a.d.SignUp(ctx, signupReq.Name, signupReq.Email, signupReq.Password)
	if signupReq.Email == "" || signupReq.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}
	return &authV2.SignupResponse{
		UserId: string(u.Id),
	}, nil
}

func (a *authGrpcServer) Login(ctx context.Context, loginRequest *authV2.LoginRequest) (*authV2.LoginResponse, error) {
	user, err := a.d.Login(ctx, loginRequest.Email, loginRequest.Password)
	if loginRequest.Email == "" || loginRequest.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	return &authV2.LoginResponse{
		UserId: string(user.Id),
	}, nil
}

func NewServer(d domain.Domain, sessionRepo kv.Repo[*common.AuthSession]) authV2.AuthV2Server {
	return &authGrpcServer{
		d:           d,
		sessionRepo: sessionRepo,
	}
}
