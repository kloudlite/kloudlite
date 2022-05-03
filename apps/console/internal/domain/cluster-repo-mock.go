package domain

import (
	"context"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

type mockClusterRepo struct {
}

var clusterIp = "64.227.166.70"
var publicKey = "YioH2JQff3T3OgJPuETlvwZKLKNK1OMoY2PBgYgvrw0="

var cluster = entities.Cluster{
	BaseEntity: repos.BaseEntity{
		Id: "dev-cluster",
	},
	AccountId:  "account-id",
	Name:       "dev-cluster",
	Provider:   "do",
	Region:     "blr1",
	Ip:         &clusterIp,
	PublicKey:  &publicKey,
	NodesCount: 3,
	Status:     "",
}

func (m mockClusterRepo) NewId() repos.ID {
	return "mock-cluster-repo"
}

func (m mockClusterRepo) Find(ctx context.Context, query repos.Query) ([]*entities.Cluster, error) {
	return []*entities.Cluster{&cluster}, nil
}

func (m mockClusterRepo) FindOne(ctx context.Context, filter repos.Filter) (*entities.Cluster, error) {
	return &cluster, nil
}

func (m mockClusterRepo) FindPaginated(ctx context.Context, query repos.Query, page int64, size int64, opts ...repos.Opts) (repos.PaginatedRecord[*entities.Cluster], error) {
	return repos.PaginatedRecord[*entities.Cluster]{}, nil
}

func (m mockClusterRepo) FindById(ctx context.Context, id repos.ID) (*entities.Cluster, error) {
	return &cluster, nil
}

func (m mockClusterRepo) Create(ctx context.Context, data *entities.Cluster) (*entities.Cluster, error) {
	return &cluster, nil
}

func (m mockClusterRepo) Upsert(ctx context.Context, filter repos.Filter, data *entities.Cluster) (*entities.Cluster, error) {
	return &cluster, nil
}

func (m mockClusterRepo) UpdateById(ctx context.Context, id repos.ID, updatedData *entities.Cluster, opts ...repos.UpdateOpts) (*entities.Cluster, error) {
	return &cluster, nil
}

func (m mockClusterRepo) DeleteById(ctx context.Context, id repos.ID) error {
	return nil
}

func (m mockClusterRepo) DeleteMany(ctx context.Context, filter repos.Filter) error {
	return nil
}

func (m mockClusterRepo) IndexFields(ctx context.Context, indices []repos.IndexField) error {
	return nil
}
