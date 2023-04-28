package app

import (
	"context"
	"kloudlite.io/apps/container-registry/internal/domain"

	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"
)

type containerRegistryGrpcServer struct {
	container_registry.UnimplementedContainerRegistryServer
	d domain.Domain
}

func (c *containerRegistryGrpcServer) CreateProjectForAccount(ctx context.Context, in *container_registry.CreateProjectIn) (*container_registry.CreateProjectOut, error) {
	return nil, nil
	// _, err := c.d.CreateHarborProject(domain.NewRegistryContext(ctx, "", in.AccountName))
	// if err != nil {
	// 	return nil, err
	// }
	// return &container_registry.CreateProjectOut{
	// 	Success: true,
	// }, nil
}

func (c *containerRegistryGrpcServer) GetSvcCredentials(ctx context.Context, in *container_registry.GetSvcCredentialsIn) (*container_registry.GetSvcCredentialsOut, error) {
	credentials, err := c.d.GetHarborCredentials(domain.NewRegistryContext(ctx, "", in.AccountName))
	if err != nil {
		return nil, err
	}
	return &container_registry.GetSvcCredentialsOut{
		UserName: credentials.Username,
		Password: credentials.Password,
	}, nil
}

func fxRPCServer(d domain.Domain) container_registry.ContainerRegistryServer {
	return &containerRegistryGrpcServer{
		d: d,
	}
}
