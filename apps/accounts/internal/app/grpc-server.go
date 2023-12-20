package app

import (
	"context"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/apps/accounts/internal/domain"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"github.com/kloudlite/api/pkg/grpc"
	"github.com/kloudlite/api/pkg/repos"
)

type AccountsGrpcServer grpc.Server

type accountsGrpcServer struct {
	accounts.UnimplementedAccountsServer
	d domain.Domain
}

// GetAccount implements accounts.AccountsServer.
func (s *accountsGrpcServer) GetAccount(ctx context.Context, in *accounts.GetAccountIn) (*accounts.GetAccountOut, error) {
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
		IsActive:        isActive,
		TargetNamespace: *acc.Spec.TargetNamespace,
		AccountId:       string(acc.Id),
	}, nil
}

func registerAccountsGRPCServer(server AccountsGrpcServer, d domain.Domain) accounts.AccountsServer {
	accountsSvc := &accountsGrpcServer{d: d}
	accounts.RegisterAccountsServer(server, accountsSvc)
	return accountsSvc
}
