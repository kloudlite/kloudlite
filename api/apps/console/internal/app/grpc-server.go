package app

import (
	"context"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/repos"
)

type grpcServer struct {
	d domain.Domain
	console.UnimplementedConsoleServer
	kcli k8s.Client
}

func (g *grpcServer) ArchiveResourcesForCluster(ctx context.Context, in *console.ArchiveResourcesForClusterIn) (*console.ArchiveResourcesForClusterOut, error) {
	consoleCtx := domain.ConsoleContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserName:    in.UserName,
		UserEmail:   in.UserEmail,
		AccountName: in.AccountName,
	}

	_, err := g.d.ArchiveEnvironmentsForCluster(consoleCtx, in.ClusterName)
	if err != nil {
		return &console.ArchiveResourcesForClusterOut{Archived: false}, err
	}

	archiveStatus, err := g.d.ArchiveClusterManagedServicesForCluster(consoleCtx, in.ClusterName)
	if err != nil {
		return &console.ArchiveResourcesForClusterOut{Archived: false}, err
	}

	return &console.ArchiveResourcesForClusterOut{Archived: archiveStatus}, nil
}

func newConsoleGrpcServer(d domain.Domain, kcli k8s.Client) console.ConsoleServer {
	return &grpcServer{
		d:    d,
		kcli: kcli,
	}
}
