package app

import (
	"context"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
)

type authGrpcServerImpl struct {
	auth.UnimplementedAuthServer
	d domain.Domain
}

func (a *authGrpcServerImpl) EnsureUserByEmail(ctx context.Context, request *auth.GetUserByEmailRequest) (*auth.GetUserByEmailOut, error) {
	user, err := a.d.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		user, err = a.d.EnsureUserByEmail(ctx, request.Email)
		if err != nil {
			return nil, err
		}
	}
	return &auth.GetUserByEmailOut{
		UserId: string(user.Id),
	}, nil
}

func (a *authGrpcServerImpl) GetAccessToken(ctx context.Context, request *auth.GetAccessTokenRequest) (*auth.AccessTokenOut, error) {
	token, err := a.d.GetAccessToken(ctx, request.Provider, request.UserId)
	if err != nil {
		return nil, err
	}
	return &auth.AccessTokenOut{
		UserId:   string(token.UserId),
		Email:    token.Email,
		Provider: token.Provider,
		OauthToken: &auth.OauthToken{
			AccessToken:  token.Token.AccessToken,
			TokenType:    token.Token.TokenType,
			RefreshToken: token.Token.RefreshToken,
			Expiry:       token.Token.Expiry.UnixMilli(),
		},
	}, err
}

func fxRPCServer(d domain.Domain) auth.AuthServer {
	return &authGrpcServerImpl{
		d: d,
	}
}
