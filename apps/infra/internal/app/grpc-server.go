package app

import (
	"context"
	"log/slog"

	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/apps/infra/protobufs/infra"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/grpc"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/repos"
	corev1 "k8s.io/api/core/v1"
)

type grpcServer struct {
	d domain.Domain
	infra.UnimplementedInfraServer
	kcli   k8s.Client
	logger *slog.Logger
}

// GetClusterGatewayResource implements infra.InfraServer.
func (g *grpcServer) GetClusterGatewayResource(ctx context.Context, in *infra.GetClusterGatewayResourceIn) (*infra.GetClusterGatewayResourceOut, error) {
	l := grpc.NewRequestLogger(g.logger, "EnsureGlobalVPNConnection")
	defer l.End()

	gw, err := g.d.GetGatewayResource(ctx, in.AccountName, in.ClusterName)
	if err != nil {
		return nil, err
	}

	b, err := fn.K8sObjToYAML(gw)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &infra.GetClusterGatewayResourceOut{Gateway: b}, nil
}

// EnsureGlobalVPNConnection implements infra.InfraServer.
func (g *grpcServer) EnsureGlobalVPNConnection(ctx context.Context, in *infra.EnsureGlobalVPNConnectionIn) (*infra.EnsureGlobalVPNConnectionOut, error) {
	l := grpc.NewRequestLogger(g.logger, "EnsureGlobalVPNConnection")
	defer l.End()
	_, err := g.d.EnsureGlobalVPNConnection(domain.InfraContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserEmail:   in.UserEmail,
		UserName:    in.UserName,
		AccountName: in.AccountName,
	}, in.ClusterName, in.GlobalVPNName, &entities.DispatchAddr{
		AccountName: in.DispatchAddr_AccountName,
		ClusterName: in.DispatchAddr_ClusterName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &infra.EnsureGlobalVPNConnectionOut{}, nil
}

// GetClusterKubeconfig implements infra.InfraServer.
func (g *grpcServer) GetClusterKubeconfig(ctx context.Context, in *infra.GetClusterIn) (*infra.GetClusterKubeconfigOut, error) {
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

	creds := &corev1.Secret{}
	if err := g.kcli.Get(ctx, fn.NN(c.Namespace, c.Spec.Output.SecretName), creds); err != nil {
		return nil, err
	}

	return &infra.GetClusterKubeconfigOut{Kubeconfig: creds.Data[c.Spec.Output.KeyKubeconfig]}, nil
}

// GetCluster implements infra.InfraServer.
func (g *grpcServer) GetByokCluster(ctx context.Context, in *infra.GetClusterIn) (*infra.GetClusterOut, error) {
	infraCtx := domain.InfraContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserEmail:   in.UserEmail,
		UserName:    in.UserName,
		AccountName: in.AccountName,
	}
	c, err := g.d.GetBYOKCluster(infraCtx, in.ClusterName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if c == nil {
		return nil, errors.Newf("cluster %s not found", in.ClusterName)
	}

	return &infra.GetClusterOut{
		OwnedBy: func() string {
			if c.OwnedBy != nil {
				return *c.OwnedBy
			}
			return ""
		}(),
	}, nil
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
		OwnedBy: func() string {
			if c.OwnedBy != nil {
				return *c.OwnedBy
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

// MarkClusterAsOnline implements infra.InfraServer.
func (g *grpcServer) MarkClusterOnlineAt(ctx context.Context, in *infra.MarkClusterOnlineAtIn) (*infra.MarkClusterOnlineAtOut, error) {
	ictx := domain.InfraContext{Context: ctx, AccountName: in.AccountName}
	if err := g.d.MarkClusterOnlineAt(ictx, in.ClusterName, fn.New(in.Timestamp.AsTime())); err != nil {
		return nil, errors.NewE(err)
	}

	return &infra.MarkClusterOnlineAtOut{}, nil
}

func newGrpcServer(d domain.Domain, kcli k8s.Client, logger *slog.Logger) infra.InfraServer {
	return &grpcServer{
		d:      d,
		kcli:   kcli,
		logger: logger,
	}
}
