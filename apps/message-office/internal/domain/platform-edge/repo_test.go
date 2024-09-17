package platform_edge

import (
	"context"
	"reflect"
	"testing"

	"github.com/kloudlite/api/apps/message-office/internal/entities"
	fc "github.com/kloudlite/api/apps/message-office/internal/entities/field-constants"
	mock_repos "github.com/kloudlite/api/mocks/pkg/repos"

	// fn "github.com/kloudlite/api/pkg/functions"
	mo_errors "github.com/kloudlite/api/apps/message-office/errors"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func TestRepo_AllocatePlatformEdgeCluster(t *testing.T) {
	type fields struct {
		edge_clusters      func(t *testing.T) repos.DbRepo[*entities.PlatformEdgeCluster]
		allocated_clusters func(t *testing.T) repos.DbRepo[*entities.ClusterAllocation]
	}
	type args struct {
		ctx     context.Context
		region  string
		account string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *entities.PlatformEdgeCluster
		wantErr error
	}{
		{
			name: "[region] not found",
			fields: fields{
				edge_clusters: func(t *testing.T) repos.DbRepo[*entities.PlatformEdgeCluster] {
					repo := mock_repos.NewDbRepo[*entities.PlatformEdgeCluster]()

					repo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.PlatformEdgeCluster, error) {
						return nil, nil
					}

					return repo
				},
				allocated_clusters: func(t *testing.T) repos.DbRepo[*entities.ClusterAllocation] {
					repo := mock_repos.NewDbRepo[*entities.ClusterAllocation]()
					repo.MockGroupByAndCount = func(ctx context.Context, filter repos.Filter, groupBy string, opts repos.GroupByAndCountOptions) (map[string]int64, error) {
						return nil, nil
					}
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				region:  "REGION",
				account: "sample-account",
			},
			want:    nil,
			wantErr: mo_errors.ErrNoClustersInRegion,
		},

		{
			name: "[region] found, but no allocation till now",
			fields: fields{
				edge_clusters: func(t *testing.T) repos.DbRepo[*entities.PlatformEdgeCluster] {
					repo := mock_repos.NewDbRepo[*entities.PlatformEdgeCluster]()

					repo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.PlatformEdgeCluster, error) {
						return &entities.PlatformEdgeCluster{
							BaseEntity:    repos.BaseEntity{Id: "1"},
							Name:          "sample-cluster",
							Region:        "REGION",
							CloudProvider: "sample-cloud-provider",
						}, nil
					}

					return repo
				},
				allocated_clusters: func(t *testing.T) repos.DbRepo[*entities.ClusterAllocation] {
					repo := mock_repos.NewDbRepo[*entities.ClusterAllocation]()

					repo.MockGroupByAndCount = func(ctx context.Context, filter repos.Filter, groupBy string, opts repos.GroupByAndCountOptions) (map[string]int64, error) {
						return nil, nil
					}

					repo.MockCreate = func(ctx context.Context, data *entities.ClusterAllocation) (*entities.ClusterAllocation, error) {
						if data.Cluster.Name != "sample-cluster" {
							t.Errorf("Repo.AllocatePlatformEdgeCluster() called with wrong cluster name")
						}
						return data, nil
					}

					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				region:  "REGION",
				account: "sample-account",
			},
			want: &entities.PlatformEdgeCluster{
				BaseEntity:    repos.BaseEntity{Id: "1"},
				Name:          "sample-cluster",
				Region:        "REGION",
				CloudProvider: "sample-cloud-provider",
			},
			wantErr: mo_errors.ErrNoClustersInRegion,
		},

		{
			name: "[allocations] found",
			fields: fields{
				edge_clusters: func(t *testing.T) repos.DbRepo[*entities.PlatformEdgeCluster] {
					repo := mock_repos.NewDbRepo[*entities.PlatformEdgeCluster]()

					repo.MockFindOne = func(ctx context.Context, filter repos.Filter) (*entities.PlatformEdgeCluster, error) {
						if filter[fc.PlatformEdgeClusterName] != "sample-cluster" {
							t.Errorf("edge_clusters.FindOne() called with wrong cluster name")
							t.FailNow()
						}

						if filter[fc.PlatformEdgeClusterRegion] != "REGION" {
							t.Errorf("edge_clusters.FindOne() called with wrong region")
							t.FailNow()
						}

						return &entities.PlatformEdgeCluster{
							BaseEntity:    repos.BaseEntity{Id: "1"},
							Name:          "sample-cluster",
							Region:        "REGION",
							CloudProvider: "sample-cloud-provider",
						}, nil
					}

					return repo
				},
				allocated_clusters: func(t *testing.T) repos.DbRepo[*entities.ClusterAllocation] {
					repo := mock_repos.NewDbRepo[*entities.ClusterAllocation]()

					repo.MockGroupByAndCount = func(ctx context.Context, filter repos.Filter, groupBy string, opts repos.GroupByAndCountOptions) (map[string]int64, error) {
						return map[string]int64{"sample-cluster": 1}, nil
					}

					repo.MockCreate = func(ctx context.Context, data *entities.ClusterAllocation) (*entities.ClusterAllocation, error) {
						if data.Cluster.Name != "sample-cluster" {
							t.Errorf("allocated_clusters.AllocatePlatformEdgeCluster() called with wrong cluster name")
						}

						if data.Cluster.Region != "REGION" {
							t.Errorf("allocated_clusters.AllocatePlatformEdgeCluster() called with wrong region")
						}

						return data, nil
					}

					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				region:  "REGION",
				account: "sample-account",
			},
			want: &entities.PlatformEdgeCluster{
				BaseEntity:    repos.BaseEntity{Id: "1"},
				Name:          "sample-cluster",
				Region:        "REGION",
				CloudProvider: "sample-cloud-provider",
			},
			wantErr: mo_errors.ErrNoClustersInRegion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Repo{
				EdgeClusters:      tt.fields.edge_clusters(t),
				AllocatedClusters: tt.fields.allocated_clusters(t),
			}
			got, err := r.AllocatePlatformEdgeCluster(tt.args.ctx, tt.args.region, tt.args.account)
			if (err != nil) && !errors.Is(err, tt.wantErr) {
				t.Errorf("Repo.AllocatePlatformEdgeCluster()\n\terror = %v,\n\twantErr = %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repo.AllocatePlatformEdgeCluster()\n\tgot = %#v,\n\twant %#v", got, tt.want)
			}
		})
	}
}
