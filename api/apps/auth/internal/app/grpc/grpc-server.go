package grpc

import (
	"context"

	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/pkg/repos"

	"github.com/kloudlite/api/apps/auth/internal/domain"
	authrpc "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/pkg/errors"
)

type authInternalGrpcServer struct {
	authrpc.UnimplementedAuthInternalServer
	d domain.Domain
	//sessionRepo kv.Repo[*common.AuthSession]
}

// GenerateMachineSession implements auth.AuthServer.
func (a *authInternalGrpcServer) GenerateMachineSession(ctx context.Context, in *authrpc.GenerateMachineSessionIn) (*authrpc.GenerateMachineSessionOut, error) {
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

func (a *authInternalGrpcServer) FromAccToken(token entities.AccessToken) *authrpc.AccessTokenOut {
	return &authrpc.AccessTokenOut{
		Id:       string(token.Id),
		UserId:   string(token.UserId),
		Email:    token.Email,
		Provider: token.Provider,
		OauthToken: &authrpc.OauthToken{
			AccessToken:  token.Token.AccessToken,
			TokenType:    token.Token.TokenType,
			RefreshToken: token.Token.RefreshToken,
			Expiry:       token.Token.Expiry.UnixMilli(),
		},
	}
}

func (a *authInternalGrpcServer) EnsureUserByEmail(ctx context.Context, request *authrpc.GetUserByEmailRequest) (*authrpc.GetUserByEmailOut, error) {
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
	return &authrpc.GetUserByEmailOut{
		UserId: string(user.Id),
	}, nil
}

func NewInternalServer(d domain.Domain) authrpc.AuthInternalServer {
	return &authInternalGrpcServer{
		d: d,
	}
}
