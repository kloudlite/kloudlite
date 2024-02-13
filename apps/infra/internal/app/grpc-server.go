package app

import (
	"context"

	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/infra"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

type grpcServer struct {
	d domain.Domain
	infra.UnimplementedInfraServer
}

func (g *grpcServer) contextFromAccount(ctx context.Context, accountName string) domain.InfraContext {
	dctx := domain.InfraContext{
		Context:     ctx,
		UserId:      "sys-user:error-on-apply-worker",
		UserEmail:   "",
		UserName:    "",
		AccountName: accountName,
	}
	return dctx
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
		return nil, errors.NewE(err)
	}

	if c == nil {
		return nil, errors.Newf("cluster %s not found", in.ClusterName)
	}

	return &infra.GetClusterOut{
		MessageQueueTopic: c.Spec.MessageQueueTopicName,
		DnsHost:           c.Spec.PublicDNSHost,

		IACJobName: func() string {
			if c.Spec.Output != nil {
				return c.Spec.Output.JobName
			}
			return ""
		}(),
		IACJobNamespace: func() string {
			if c.Spec.Output != nil {
				return c.Spec.Output.JobNamespace
			}
			return ""
		}(),
	}, nil
}

// GetNodepool implements infra.InfraServer
func (g *grpcServer) GetNodepool(ctx context.Context, in *infra.GetNodepoolIn) (*infra.GetNodepoolOut, error) {
	infraCtx := domain.InfraContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserEmail:   in.UserEmail,
		UserName:    in.UserName,
		AccountName: in.AccountName,
	}
	np, err := g.d.GetNodePool(infraCtx, in.ClusterName, in.NodepoolName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if np == nil {
		return nil, errors.Newf("nodepool %s not found", in.NodepoolName)
	}

	return &infra.GetNodepoolOut{
		// IACJobName:      np.Spec.IAC.JobName,
		// IACJobNamespace: np.Spec.IAC.JobNamespace,
	}, nil
}

func (g *grpcServer) ClusterExists(ctx context.Context, in *infra.ClusterExistsIn) (*infra.ClusterExistsOut, error) {
	infraCtx := domain.InfraContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserEmail:   in.UserEmail,
		UserName:    in.UserName,
		AccountName: in.AccountName,
	}
	cluster, err := g.d.GetCluster(infraCtx, in.ClusterName)
	if err != nil {
		if errors.Is(err, domain.ErrClusterNotFound) {
			return &infra.ClusterExistsOut{Exists: false}, nil
		}
		return nil, errors.NewE(err)
	}

	if cluster == nil {
		return &infra.ClusterExistsOut{Exists: false}, nil
	}

	return &infra.ClusterExistsOut{Exists: true}, nil
}

func newGrpcServer(d domain.Domain) infra.InfraServer {
	return &grpcServer{
		d: d,
	}
}
