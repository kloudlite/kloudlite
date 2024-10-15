package platform_edge

import (
	"context"

	"github.com/kloudlite/api/apps/message-office/internal/entities"
)

type Domain interface {
	AllocatePlatformEdgeCluster(ctx context.Context, region string, account string) (*entities.PlatformEdgeCluster, error)
	GetAllocatedPlatformEdgeCluster(ctx context.Context, account string) (*entities.ClusterAllocation, error)
	ListPlatformEdgeClusters(ctx context.Context, region *string) ([]*entities.PlatformEdgeCluster, error)
}

var _ Domain = (*Repo)(nil)
