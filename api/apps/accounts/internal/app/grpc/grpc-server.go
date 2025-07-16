package grpc

import (
	"context"
	"github.com/kloudlite/api/apps/accounts/internal/domain"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

type accountsInternalGrpcServer struct {
	accounts.UnimplementedAccountsInternalServer
	d domain.Domain
}

func (s *accountsInternalGrpcServer) GetAccount(ctx context.Context, in *accounts.GetAccountIn) (*accounts.GetAccountOut, error) {
	acc, err := s.d.GetAccount(domain.UserContext{
		Context: ctx,
		UserId:  repos.ID(in.UserId),
	}, in.AccountName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	isActive := false
	if acc.IsActive != nil {
		isActive = *acc.IsActive
	}

	return &accounts.GetAccountOut{
		IsActive:               isActive,
		TargetNamespace:        acc.TargetNamespace,
		AccountId:              string(acc.Id),
		KloudliteGatewayRegion: acc.KloudliteGatewayRegion,
	}, nil
}

func NewAccountsInternalServer(d domain.Domain) accounts.AccountsInternalServer {
	serverImpl := &accountsInternalGrpcServer{d: d}
	return serverImpl
}
