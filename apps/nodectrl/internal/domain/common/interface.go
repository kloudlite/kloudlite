package common

import "context"

type ProviderClient interface {
	NewNode(ctx context.Context) error
	DeleteNode(ctx context.Context) error

	SaveToDbGuranteed(ctx context.Context)
}
