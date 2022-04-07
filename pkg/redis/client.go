package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type redisCli[T any] struct {
	client     *redis.Client
	clientOpts *redis.Options
}

func (r *redisCli[T]) Set(ctx context.Context, key string, value T) (T, error) {
	r.client.Set(ctx, key, value)
}

func (r *redisCli[T]) Get(key string) T {
	panic("not implemented") // TODO: Implement
}

func (r *redisCli[T]) Expire(key string) bool {
	panic("not implemented") // TODO: Implement
}

func (r *redisCli[T]) Drop(key string) bool {
	panic("not implemented") // TODO: Implement
}

func (r *redisCli[T]) Connect(ctx context.Context) error {
	r.client = redis.NewClient(r.clientOpts)
	return r.client.Ping(ctx).Err()
}

type RedisClient[T any] interface {
	Connect(ctx context.Context) error
	Set(key string, value T) (T, error)
	Get(key string) T
	Expire(key string) bool
	Drop(key string) bool
}

func NewClient[T any](opts *redis.Options) RedisClient[T] {
	return &redisCli[T]{client: nil, clientOpts: opts}
}
