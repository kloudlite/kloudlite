package cache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"
)

type redisRepo[T any] struct {
	opts struct {
		Addr     string
		Username string
		Password string
	}
	client *redis.Client
}

func (r *redisRepo[T]) Connect(context.Context) error {
	r.client = redis.NewClient(&redis.Options{
		Addr:     r.opts.Addr,
		Password: r.opts.Password,
		Username: r.opts.Username,
	})
	return nil
}

func (r *redisRepo[T]) Close(context.Context) error {
	err := r.client.Close()
	if err != nil {
		return err
	}
	r.client = nil
	return nil
}

func (r *redisRepo[T]) Set(c context.Context, key string, value T) error {
	marshal, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(c, key, string(marshal), 0).Err()
}

func (r *redisRepo[T]) SetWithExpiry(c context.Context, key string, value T, duration time.Duration) error {
	marshal, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(c, key, string(marshal), duration).Err()
}

func (r *redisRepo[T]) Get(c context.Context, key string) (T, error) {
	result, err := r.client.Get(c, key).Result()
	if err != nil {
		return nil, err
	}
	var value T
	err = json.Unmarshal([]byte(result), &value)
	return value, err
}

func (r *redisRepo[T]) Drop(c context.Context, key string) error {
	return r.client.Del(c, key).Err()
}

func NewRedisRepo[T any](addr string, password string, userName string) Repo[T] {
	return &redisRepo[T]{
		opts: struct {
			Addr     string
			Username string
			Password string
		}{Addr: addr, Username: userName, Password: password},
	}
}
