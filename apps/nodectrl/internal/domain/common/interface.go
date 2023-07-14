package common

import "context"

type ProviderClient interface {
	NewNode(ctx context.Context) error
	DeleteNode(ctx context.Context, force bool) error
	SaveToDbGuranteed(ctx context.Context)

	CreateCluster(ctx context.Context) error

	AddWorker(ctx context.Context) error
	AddMaster(ctx context.Context) error
}
