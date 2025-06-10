package graphv2

import (
	"context"

	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/common"
	authV2 "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth/v2"
	"github.com/kloudlite/api/pkg/kv"
)

type authGrpcServer struct {
	authV2.UnimplementedAuthV2Server
	d           domain.Domain
	sessionRepo kv.Repo[*common.AuthSession]
}

func (a *authGrpcServer) Login(ctx context.Context, loginRequest *authV2.LoginRequest) (*authV2.LoginResponse, error) {
	user, err := a.d.Login(ctx, loginRequest.Email, loginRequest.Password)
	if err != nil {
		return nil, err
	}
	return &authV2.LoginResponse{
		UserId: string(user.Id),
	}, nil
}

func NewAuthGrpcServer(d domain.Domain, sessionRepo kv.Repo[*common.AuthSession]) authV2.AuthV2Server {
	return &authGrpcServer{
		d:           d,
		sessionRepo: sessionRepo,
	}
}
