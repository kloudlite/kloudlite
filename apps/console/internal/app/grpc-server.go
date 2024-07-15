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

func (g *grpcServer) ArchiveEnvironmentsForCluster(ctx context.Context, in *console.ArchiveEnvironmentsForClusterIn) (*console.ArchiveEnvironmentsForClusterOut, error) {
	consoleCtx := domain.ConsoleContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserName:    in.UserName,
		UserEmail:   in.UserEmail,
		AccountName: in.AccountName,
	}

	archiveStatus, err := g.d.ArchiveEnvironmentsForCluster(consoleCtx, in.ClusterName)
	if err != nil {
		return &console.ArchiveEnvironmentsForClusterOut{Archived: false}, err
	}

	return &console.ArchiveEnvironmentsForClusterOut{Archived: archiveStatus}, nil
}

func newConsoleGrpcServer(d domain.Domain, kcli k8s.Client) console.ConsoleServer {
	return &grpcServer{
		d:    d,
		kcli: kcli,
	}
}
