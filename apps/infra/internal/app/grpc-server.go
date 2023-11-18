package app

import (
	"context"

	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
	"kloudlite.io/pkg/repos"
)

type grpcServer struct {
	d domain.Domain
	infra.UnimplementedInfraServer
}

// GetCluster implements infra.InfraServer.
func (g *grpcServer) GetCluster(ctx context.Context, in *infra.GetClusterIn) (*infra.GetClusterOut, error) {
	infraCtx := domain.InfraContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserEmail:   in.UserEmail,
		UserName:    in.UserName,
		AccountName: in.AccountName,
	}
	c, err := g.d.GetCluster(infraCtx, in.ClusterName)
	if err != nil {
		return nil, err
	}

	return &infra.GetClusterOut{
		MessageQueueTopic: c.Spec.MessageQueueTopicName,
		DnsHost:           c.Spec.PublicDNSHost,
	}, nil
}

func newGrpcServer(d domain.Domain) *grpcServer {
	return &grpcServer{
		d: d,
	}
}
