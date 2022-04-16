package cache

import (
	"context"
	"time"
)

type Client interface {
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	Set(c context.Context, key string, value []byte) error
	SetWithExpiry(c context.Context, key string, value []byte, duration time.Duration) error
	Drop(c context.Context, key string) error
	Get(c context.Context, key string) ([]byte, error)
}

type Repo[T any] interface {
	Set(c context.Context, key string, value T) error
	SetWithExpiry(c context.Context, key string, value T, duration time.Duration) error
	Get(c context.Context, key string) (T, error)
	Drop(c context.Context, key string) error
}
