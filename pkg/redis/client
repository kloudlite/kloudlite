package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kloudlite/api/pkg/errors"
)

type redisCli[T any] struct {
	client     *redis.Client
	clientOpts *redis.Options
}

func (r *redisCli[T]) Get(ctx context.Context, key string) (*T, error) {
	var result T
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, errors.NewEf(err, "could not get key (%s)")
	}
	err = json.Unmarshal([]byte(val), &result)
	if err != nil {
		return nil, errors.NewEf(err, "could not unmarshal value into type (%T)", result)
	}
	return &result, nil
}

func (r *redisCli[T]) Expire(ctx context.Context, key string, after time.Duration) (bool, error) {
	return r.client.Expire(ctx, key, after).Result()
}

func (r *redisCli[T]) Drop(ctx context.Context, key string) (bool, error) {
	v, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return false, errors.NewEf(err, "could not delete key (%s)", key)
	}
	return v > 0, nil
}

func (r *redisCli[T]) Set(ctx context.Context, key string, value T) error {
	err := r.client.Set(ctx, key, value, time.Duration(0)).Err()
	if err != nil {
		return errors.NewEf(err, "could not set key (%s)", key)
	}
	return nil
}

func (r *redisCli[T]) SetE(ctx context.Context, key string, value T, expireAfter time.Duration) error {
	err := r.client.Set(ctx, key, value, expireAfter).Err()
	if err != nil {
		return errors.NewEf(err, "could not set key (%s)", key)
	}
	return nil
}

func (r *redisCli[T]) Connect(ctx context.Context) error {
	r.client = redis.NewClient(r.clientOpts)
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return errors.NewEf(err, "could not connect to redis")
	}
	return nil
}

type RedisClient[T any] interface {
	Connect(ctx context.Context) error
	Set(ctx context.Context, key string, value T) error
	Get(ctx context.Context, key string) (*T, error)
	Expire(ctx context.Context, key string, after time.Duration) (bool, error)
	Drop(ctx context.Context, key string) (bool, error)
}

func NewClient[T any](opts *redis.Options) RedisClient[T] {
	return &redisCli[T]{client: nil, clientOpts: opts}
}
