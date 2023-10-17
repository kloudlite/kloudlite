package domain

import (
	"context"
)

type Domain interface {
	ProcessRegistryEvents(ctx context.Context) error
}
