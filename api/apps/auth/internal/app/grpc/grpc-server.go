package grpc

import (
	"context"
	googleGrpc "google.golang.org/grpc"

	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/pkg/repos"

	"github.com/kloudlite/api/apps/auth/internal/domain"
	auth_rpc "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/pkg/errors"
)

type authInternalGrpcServer struct {
	auth_rpc.UnimplementedAuthInternalServer
	d domain.Domain
	//sessionRepo kv.Repo[*common.AuthSession]
}

// GenerateMachineSession implements auth.AuthServer.
func (a *authInternalGrpcServer) GenerateMachineSession(ctx context.Context, in *auth_rpc.GenerateMachineSessionIn) (*auth_rpc.GenerateMachineSessionOut, error) {
	//session, err := a.d.MachineLogin(ctx, in.UserId, in.MachineId, in.Cluster)
	//if err != nil {
	//	return nil, errors.NewE(err)
	//}
	//err = a.sessionRepo.Set(ctx, string(session.Id), session)
	//if err != nil {
	//	return nil, errors.NewE(err)
	//}
	//if session == nil {
	//	return nil, errors.Newf("session is nil")
	//}
	//return &auth.GenerateMachineSessionOut{
	//	SessionId: string(session.Id),
	//}, nil
	return nil, nil
}

// ClearMachineSessionByMachine implements auth.AuthServer.
func (a *authInternalGrpcServer) ClearMachineSessionByMachine(context.Context, *auth_rpc.ClearMachineSessionByMachineIn) (*auth_rpc.ClearMachineSessionByMachineOut, error) {
	panic("unimplemented")
}

// ClearMachineSessionByTeam implements auth.AuthServer.
func (a *authInternalGrpcServer) ClearMachineSessionByTeam(context.Context, *auth_rpc.ClearMachineSessionByTeamIn) (*auth_rpc.ClearMachineSessionByTeamOut, error) {
	panic("unimplemented")
}

// ClearMachineSessionByUser implements auth.AuthServer.
func (a *authInternalGrpcServer) ClearMachineSessionByUser(context.Context, *auth_rpc.ClearMachineSessionByUserIn) (*auth_rpc.ClearMachineSessionByUserOut, error) {
	panic("unimplemented")
}

func (a *authInternalGrpcServer) GetUser(ctx context.Context, in *auth_rpc.GetUserIn) (*auth_rpc.GetUserOut, error) {
	user, err := a.d.GetUserById(ctx, repos.ID(in.UserId))
	if err != nil {
		return nil, errors.NewE(err)
	}
	if user == nil {
		return nil, errors.Newf("could not find user with (id=%q)", in.UserId)
	}
	return &auth_rpc.GetUserOut{
		Id:    string(user.Id),
		Email: user.Email,
		Name:  user.Name,
	}, nil
}

func (a *authInternalGrpcServer) FromAccToken(token entities.AccessToken) *auth_rpc.AccessTokenOut {
	return &auth_rpc.AccessTokenOut{
		Id:       string(token.Id),
		UserId:   string(token.UserId),
		Email:    token.Email,
		Provider: token.Provider,
		OauthToken: &auth_rpc.OauthToken{
			AccessToken:  token.Token.AccessToken,
			TokenType:    token.Token.TokenType,
			RefreshToken: token.Token.RefreshToken,
			Expiry:       token.Token.Expiry.UnixMilli(),
		},
	}
}

func (a *authInternalGrpcServer) EnsureUserByEmail(ctx context.Context, request *auth_rpc.GetUserByEmailRequest) (*auth_rpc.GetUserByEmailOut, error) {
	user, err := a.d.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if user == nil {
		user, err = a.d.EnsureUserByEmail(ctx, request.Email)
		if err != nil {
			return nil, errors.NewE(err)
		}
	}
	return &auth_rpc.GetUserByEmailOut{
		UserId: string(user.Id),
	}, nil
}

func NewInternalServer(grpcServer *googleGrpc.Server, d domain.Domain) auth_rpc.AuthInternalServer {
	serverImpl := &authInternalGrpcServer{
		d: d,
	}
	auth_rpc.RegisterAuthInternalServer(grpcServer, serverImpl)
	return serverImpl
}
