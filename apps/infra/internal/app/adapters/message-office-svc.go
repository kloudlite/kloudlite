package adapters

import (
	"context"

	"github.com/kloudlite/api/apps/infra/internal/domain/ports"
	cluster_token "github.com/kloudlite/api/apps/message-office/protobufs/cluster-token"
	platform_edge "github.com/kloudlite/api/apps/message-office/protobufs/platform-edge"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/grpc"
	"go.uber.org/fx"
)

type MesasageOfficeGRPCClient grpc.Client

type MessageOfficeService struct {
	clusterTokenClient cluster_token.ClusterTokenClient
	platformEdgeClient platform_edge.PlatformEdgeClient
}

// GenerateClusterToken implements ports.MessageOfficeService.
func (m *MessageOfficeService) GenerateClusterToken(ctx context.Context, in *ports.GenerateClusterTokenIn) (*ports.GenerateClusterTokenOut, error) {
	tout, err := m.clusterTokenClient.GenerateClusterToken(ctx, &cluster_token.GenerateClusterTokenIn{
		AccountName: in.AccountName,
		ClusterName: in.ClusterName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &ports.GenerateClusterTokenOut{ClusterToken: tout.ClusterToken}, nil
}

// GetAllocatedPlatformEdgeCluster implements ports.MessageOfficeService.
func (m *MessageOfficeService) GetAllocatedPlatformEdgeCluster(ctx context.Context, args *ports.GetAllocatedPlatformEdgeClusterIn) (*ports.GetAllocatedPlatformEdgeClusterOut, error) {
	out, err := m.platformEdgeClient.GetAllocatedPlatformEdgeCluster(ctx, &platform_edge.GetAllocatedPlatformEdgeClusterIn{
		AccountName: args.AccountName,
		ClusterName: args.ClusterName,
	})
	if err != nil {
		return nil, err
	}

	return &ports.GetAllocatedPlatformEdgeClusterOut{
		PublicDNSHost: out.GetPublicDnsHost(),
	}, nil
}

// GetClusterToken implements ports.MessageOfficeService.
func (m *MessageOfficeService) GetClusterToken(ctx context.Context, in *ports.GetClusterTokenIn) (*ports.GetClusterTokenOut, error) {
	tout, err := m.clusterTokenClient.GetClusterToken(ctx, &cluster_token.GetClusterTokenIn{
		AccountName: in.AccountName,
		ClusterName: in.ClusterName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &ports.GetClusterTokenOut{ClusterToken: tout.ClusterToken}, nil
}

var _ ports.MessageOfficeService = (*MessageOfficeService)(nil)

func NewMessageOfficeService(client MesasageOfficeGRPCClient) ports.MessageOfficeService {
	return &MessageOfficeService{
		clusterTokenClient: cluster_token.NewClusterTokenClient(client),
		platformEdgeClient: platform_edge.NewPlatformEdgeClient(client),
	}
}

func FxNewMessageOfficeService() fx.Option {
	return fx.Module("message_office_service",
		fx.Provide(func(client MesasageOfficeGRPCClient) ports.MessageOfficeService {
			return NewMessageOfficeService(client)
		}),
		fx.Invoke(func(lf fx.Lifecycle, client MesasageOfficeGRPCClient) {
			lf.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					return client.Close()
				},
			})
		}),
	)
}
