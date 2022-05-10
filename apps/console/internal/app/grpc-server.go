package app

import (
	"context"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/pkg/repos"
)

type consoleServerI struct {
	console.UnimplementedConsoleServer
	d domain.Domain
}

func (c *consoleServerI) SetupClusterForAccount(ctx context.Context, in *console.AccountIn) (*console.SetupClusterVoid, error) {
	_, err := c.d.CreateClusterAccount(ctx, &entities.ClusterAccount{
		AccountID: repos.ID(in.AccountId),
	}, in.Region, in.Provider)
	if err != nil {
		return nil, err
	}
	return &console.SetupClusterVoid{}, nil
}

func fxConsoleGrpcServer(d domain.Domain) console.ConsoleServer {
	return &consoleServerI{
		d: d,
	}
}
