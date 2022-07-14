package app

import (
	"context"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/pkg/repos"
)

type consoleServerI struct {
	console.UnimplementedConsoleServer
	d domain.Domain
}

func (c *consoleServerI) GetProjectName(ctx context.Context, in *console.ProjectIn) (*console.ProjectOut, error) {
	project, err := c.d.GetProjectWithID(ctx, repos.ID(in.ProjectId))
	if err != nil {
		return nil, err
	}
	return &console.ProjectOut{Name: project.Name}, nil
}

func fxConsoleGrpcServer(d domain.Domain) console.ConsoleServer {
	return &consoleServerI{
		d: d,
	}
}
