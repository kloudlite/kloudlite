package app

import (
	"context"

	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/grpc-interfaces/infra"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/repos"
	corev1 "k8s.io/api/core/v1"
)

type grpcServer struct {
	d domain.Domain
	infra.UnimplementedInfraServer
	kcli k8s.Client
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

func (g *grpcServer) GetClusterManagedService(ctx context.Context, in *infra.GetClusterManagedServiceIn) (*infra.GetClusterManagedServiceOut, error) {
	infraCtx := domain.InfraContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserEmail:   in.UserEmail,
		UserName:    in.UserName,
		AccountName: in.AccountName,
	}
	msvc, err := g.d.GetClusterManagedService(infraCtx, in.MsvcName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if msvc == nil {
		return nil, errors.Newf("cluster managed service %s not found", in.MsvcName)
	}

	return &infra.GetClusterManagedServiceOut{
		TargetNamespace: msvc.Spec.TargetNamespace,
		ClusterName:     msvc.ClusterName,
	}, nil
}

func newGrpcServer(d domain.Domain, kcli k8s.Client) infra.InfraServer {
	return &grpcServer{
		d:    d,
		kcli: kcli,
	}
}
