package app

import (
	"context"
	// "encoding/json"
	"fmt"

	// "google.golang.org/protobuf/types/known/anypb"
	"kloudlite.io/apps/consolev2.old/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/pkg/repos"
)

type consoleServerI struct {
	console.UnimplementedConsoleServer
	d domain.Domain
}

func (c *consoleServerI) GetApp(ctx context.Context, in *console.AppIn) (*console.AppOut, error) {
	return nil, fmt.Errorf("Not implemented")
	// app, err := c.d.GetApp(ctx, repos.ID(in.AppId))
	// if err != nil {
	// 	return nil, err
	// }
	// marshal, err := json.Marshal(app)
	// if err != nil {
	// 	return nil, err
	// }
	// return &console.AppOut{
	// 	Data: &anypb.Any{
	// 		Value: marshal,
	// 	},
	// }, nil
}

func (c *consoleServerI) GetManagedSvc(ctx context.Context, in *console.MSvcIn) (*console.MSvcOut, error) {
	return nil, fmt.Errorf("Not implemented")
	// msvc, err := c.d.GetManagedSvc(ctx, repos.ID(in.MsvcId))
	// if err != nil {
	// 	return nil, err
	// }
	// marshal, err := json.Marshal(msvc)
	// if err != nil {
	// 	return nil, err
	// }
	// return &console.MSvcOut{
	// 	Data: &anypb.Any{
	// 		Value: marshal,
	// 	},
	// }, nil
}

func (c *consoleServerI) SetupAccount(ctx context.Context, in *console.AccountSetupIn) (*console.AccountSetupVoid, error) {
	_, err := c.d.SetupAccount(ctx, repos.ID(in.AccountId))
	if err != nil {
		return nil, err
	}
	return &console.AccountSetupVoid{}, nil
}

func (c *consoleServerI) GetProjectName(ctx context.Context, in *console.ProjectIn) (*console.ProjectOut, error) {
	return nil, fmt.Errorf("Not implemented")
	// project, err := c.d.GetProjectWithID(ctx, repos.ID(in.ProjectId))
	// if err != nil {
	// 	return nil, err
	// }
	// return &console.ProjectOut{Name: project.Name}, nil
}

func fxConsoleGrpcServer(d domain.Domain) console.ConsoleServer {
	return &consoleServerI{
		d: d,
	}
}
