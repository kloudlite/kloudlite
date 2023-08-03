package app

import (
	"context"

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
	a, err := server.d.GetAccount(ctx, req.AccountName)
	if err != nil || a == nil {
		return &accounts.CheckAccountExistsResponse{Exists: false, IsActive: false}, err
	}
	return &accounts.CheckAccountExistsResponse{Exists: a != nil, IsActive: *a.IsActive}, nil
}
