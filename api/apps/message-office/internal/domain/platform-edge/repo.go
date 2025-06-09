package platform_edge

import (
	"context"

	mo_errors "github.com/kloudlite/api/apps/message-office/errors"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/apps/message-office/internal/entities"
	fc "github.com/kloudlite/api/apps/message-office/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/repos"
)

type Repo struct {
	EdgeClusters      repos.DbRepo[*entities.PlatformEdgeCluster]
	AllocatedClusters repos.DbRepo[*entities.ClusterAllocation]
}

func (r *Repo) ListPlatformEdgeClusters(ctx context.Context, region *string) ([]*entities.PlatformEdgeCluster, error) {
	q := repos.Query{}
	if region != nil {
		q.Filter = repos.Filter{
			fc.PlatformEdgeClusterRegion: region,
		}
	}
	return r.EdgeClusters.Find(ctx, q)
}

func (r *Repo) AllocatePlatformEdgeCluster(ctx context.Context, region string, account string) (*entities.PlatformEdgeCluster, error) {
	m, err := r.AllocatedClusters.GroupByAndCount(ctx, repos.Filter{fc.ClusterAllocationClusterRegion: region}, fc.ClusterAllocationClusterName, repos.GroupByAndCountOptions{
		Limit: 1,
		Sort:  repos.SortDirectionAsc,
	})
	if err != nil {
		return nil, err
	}

	if len(m) > 1 {
		return nil, errors.New("more than one cluster available")
	}

	var clusterName string

	switch len(m) {
	case 0:
		{
			x, err := r.EdgeClusters.FindOne(ctx, repos.Filter{fc.PlatformEdgeClusterRegion: region})
			if err != nil {
				return nil, err
			}
			if x == nil {
				return nil, mo_errors.ErrNoClustersInRegion
			}

			clusterName = x.Name

			if _, err := r.AllocatedClusters.Create(ctx, &entities.ClusterAllocation{
				To: account,
				Cluster: entities.ClusterAllocationClusterRef{
					Name:           clusterName,
					Region:         region,
					OwnedByAccount: x.OwnedByAccount,
					PublicDNSHost:  x.PublicDNSHostname,
				},
			}); err != nil {
				return nil, err
			}
		}
	case 1:
		{
			for k := range m {
				clusterName = k
				break
			}
		}
	}

	pec, err := r.EdgeClusters.FindOne(ctx, repos.Filter{
		fc.PlatformEdgeClusterName:   clusterName,
		fc.PlatformEdgeClusterRegion: region,
	})
	if err != nil {
		return nil, err
	}

	if pec == nil {
		return nil, mo_errors.ErrNoClustersInRegion
	}

	return pec, nil
}

func (r *Repo) GetPlatformEdgeCluster(ctx context.Context, clusterName string) (*entities.PlatformEdgeCluster, error) {
	pec, err := r.EdgeClusters.FindOne(ctx, repos.Filter{
		fc.PlatformEdgeClusterName: clusterName,
	})
	if err != nil {
		return nil, err
	}

	if pec == nil {
		return nil, errors.New("cluster not found")
	}

	return pec, nil
}

func (r *Repo) GetAllocatedPlatformEdgeCluster(ctx context.Context, account string) (*entities.ClusterAllocation, error) {
	rec, err := r.AllocatedClusters.FindOne(ctx, repos.Filter{
		fc.ClusterAllocationTo: account,
	})
	if err != nil {
		return nil, err
	}

	if rec == nil {
		return nil, mo_errors.ErrEdgeClusterNotAllocated
	}

	return rec, nil
}
