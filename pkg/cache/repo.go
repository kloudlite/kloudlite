package cache

import (
	"context"
	"encoding/json"
	"kloudlite.io/pkg/errors"
	"time"
)

type redisRepo[T any] struct {
	cli Client
}

func (r *redisRepo[T]) Set(c context.Context, key string, value T) error {
	b, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "failed to marshal value")
	}
	return r.cli.Set(c, key, b)
}

func (r *redisRepo[T]) Get(c context.Context, key string) (*T, error) {
	get, err := r.cli.Get(c, key)
	if err != nil {
		return nil, err
	}
	var value T
	err = json.Unmarshal(get, &value)
	return &value, err
}

func (r *redisRepo[T]) SetWithExpiry(c context.Context, key string, value T, duration time.Duration) error {
	marshal, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = r.cli.SetWithExpiry(c, key, marshal, duration)
	if err != nil {
		return errors.NewEf(err, "coult not set key (%s)", key)
	}
	return nil
}

func (r *redisRepo[T]) Drop(c context.Context, key string) error {
	return r.cli.Drop(c, key)
}

func NewRepo[T any](cli Client) Repo[T] {
	return &redisRepo[T]{
		cli,
	}
}
