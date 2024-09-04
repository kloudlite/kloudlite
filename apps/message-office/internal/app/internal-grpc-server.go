package app

import (
	"context"

	"github.com/kloudlite/api/apps/message-office/internal/domain"
	cluster_token "github.com/kloudlite/api/apps/message-office/protobufs/cluster-token"
	platform_edge "github.com/kloudlite/api/apps/message-office/protobufs/platform-edge"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/grpc"
)

type (
	InternalGrpcServer grpc.Server
)

type internalServer struct {
	d domain.Domain
	// message_office_internal.UnimplementedMessageOfficeInternalServer
	cluster_token.UnimplementedClusterTokenServer
	platform_edge.UnimplementedPlatformEdgeServer
}

// GenerateClusterToken implements cluster_token.ClusterTokenServer.
func (s *internalServer) GenerateClusterToken(ctx context.Context, in *cluster_token.GenerateClusterTokenIn) (*cluster_token.GenerateClusterTokenOut, error) {
	token, err := s.d.GenClusterToken(ctx, in.AccountName, in.ClusterName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &cluster_token.GenerateClusterTokenOut{ClusterToken: token}, nil
}

// GetClusterToken implements cluster_token.ClusterTokenServer.
func (s *internalServer) GetClusterToken(ctx context.Context, in *cluster_token.GetClusterTokenIn) (*cluster_token.GetClusterTokenOut, error) {
	token, err := s.d.GetClusterToken(ctx, in.AccountName, in.ClusterName)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return &cluster_token.GetClusterTokenOut{ClusterToken: token}, nil
}

// AllocatePlatformEdgeCluster implements platform_edge.PlatformEdgeServer.
func (s *internalServer) AllocatePlatformEdgeCluster(ctx context.Context, in *platform_edge.AllocatePlatformEdgeClusterIn) (*platform_edge.AllocatePlatformEdgeClusterOut, error) {
	pec, err := s.d.AllocatePlatformEdgeCluster(ctx, in.Region, in.AccountName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &platform_edge.AllocatePlatformEdgeClusterOut{ClusterName: pec.Name}, nil
}

// GetAllocatedPlatformEdgeCluster implements platform_edge.PlatformEdgeServer.
func (s *internalServer) GetAllocatedPlatformEdgeCluster(ctx context.Context, in *platform_edge.GetAllocatedPlatformEdgeClusterIn) (*platform_edge.AllocatePlatformEdgeClusterOut, error) {
	allocated, err := s.d.GetAllocatedPlatformEdgeCluster(ctx, in.AccountName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &platform_edge.AllocatePlatformEdgeClusterOut{
		ClusterName:    allocated.Cluster.Name,
		OwnedByAccount: allocated.Cluster.OwnedByAccount,
	}, nil
}

// ListPlatformEdgeClusters implements platform_edge.PlatformEdgeServer.
func (s *internalServer) ListPlatformEdgeClusters(ctx context.Context, in *platform_edge.ListPlatformEdgeClustersIn) (*platform_edge.ListPlatformEdgeClustersOut, error) {
	pec, err := s.d.ListPlatformEdgeClusters(ctx, &in.Region)
	if err != nil {
		return nil, errors.NewE(err)
	}

	out := &platform_edge.ListPlatformEdgeClustersOut{
		PlatformEdgeClusters: make([]*platform_edge.PlatformEdgeCluster, 0, len(pec)),
	}

	for _, pec := range pec {
		out.PlatformEdgeClusters = append(out.PlatformEdgeClusters, &platform_edge.PlatformEdgeCluster{
			Region:      pec.Region,
			ClusterName: pec.Name,
		})
	}
	return out, nil
}

func newInternalMsgServer(d domain.Domain) (cluster_token.ClusterTokenServer, platform_edge.PlatformEdgeServer) {
	server := &internalServer{d: d}
	return server, server
}

func RegisterInternalMsgServer(server InternalGrpcServer, cts cluster_token.ClusterTokenServer, pes platform_edge.PlatformEdgeServer) {
	cluster_token.RegisterClusterTokenServer(server, cts)
	platform_edge.RegisterPlatformEdgeServer(server, pes)
}
