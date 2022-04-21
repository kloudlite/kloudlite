package app

import (
	"context"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
)

type consoleServerI struct {
	console.UnimplementedConsoleServer
	d domain.Domain
}

func (c consoleServerI) CreateDefaultCluster(ctx context.Context, in *console.CreateClusterIn) (*console.CreateClusterOut, error) {
	cluster, err := c.d.CreateCluster(ctx, &entities.Cluster{
		AccountId:  in.AccountId,
		Name:       in.AccountName,
		Provider:   "do",
		Region:     "blr1",
		NodesCount: 2,
	})
	return &console.CreateClusterOut{
		ClusterId: string(cluster.Id),
	}, err
}

func fxConsoleGrpcServer(d domain.Domain) console.ConsoleServer {
	return &consoleServerI{
		d: d,
	}
}
