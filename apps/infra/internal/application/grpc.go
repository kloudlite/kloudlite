package application

import (
	"context"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
	"kloudlite.io/pkg/repos"
)

type InfraGrpcSvc struct {
	infra.UnimplementedInfraServer
	d domain.Domain
}

func (i InfraGrpcSvc) GetResourceOutput(ctx context.Context, input *infra.GetInput) (*infra.Output, error) {
	output, err := i.d.GetResourceOutput(
		ctx,
		repos.ID(input.ClusterId),
		input.ManagedResName,
		input.Namespace,
	)
	if err != nil {
		return nil, err
	}
	return &infra.Output{
		Output: output,
	}, nil
}

func fxInfraGrpcServer(d domain.Domain) infra.InfraServer {
	return &InfraGrpcSvc{
		d: d,
	}
}
