package app

import (
	"context"
	"kloudlite.io/pkg/repos"

	"kloudlite.io/apps/accounts/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/accounts"
	"kloudlite.io/pkg/grpc"
)

type AccountsGrpcServer grpc.Server

type accountsGrpcServer struct {
	accounts.UnimplementedAccountsServer
	d domain.Domain
}

// CheckAccountExists implements accounts.AccountsServer.
func (server *accountsGrpcServer) CheckAccountExists(ctx context.Context, req *accounts.CheckAccountExistsRequest) (*accounts.CheckAccountExistsResponse, error) {
	acc, err := server.d.GetAccount(domain.AccountsContext{Context: ctx, AccountName: req.AccountName, UserId: repos.ID(req.UserId)}, req.AccountName)
	if err != nil {
		return nil, err
	}

	if acc == nil {
		return &accounts.CheckAccountExistsResponse{Result: false}, nil
	}

	return &accounts.CheckAccountExistsResponse{Result: acc.IsActive != nil && *acc.IsActive == true}, nil
}
