package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"kloudlite.io/pkg/errors"
)

type redisRepo[T any] struct {
	*RedisClient
}

func (r *redisRepo[T]) Set(c context.Context, key string, value T) error {
	marshal, err := json.Marshal(value)
	if err != nil {
		return errors.NewEf(err, "could not unmarshal T into JSON string")
	}
	err = r.client.Set(c, key, string(marshal), 0).Err()
	if err != nil {
		return errors.NewEf(err, "coult not set key (%s)", key)
	}
	return nil
}

func (r *redisRepo[T]) SetWithExpiry(c context.Context, key string, value T, duration time.Duration) error {
	marshal, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = r.client.Set(c, key, string(marshal), duration).Err()
	if err != nil {
		return errors.NewEf(err, "coult not set key (%s)", key)
	}
	return nil
}

func (r *redisRepo[T]) Get(c context.Context, key string) (*T, error) {
	result, err := r.client.Get(c, key).Result()
	if err != nil {
		return nil, errors.NewEf(err, "could not get key (%s)", key)
	}
	var value T
	err = json.Unmarshal([]byte(result), &value)
	if err != nil {
		return nil, errors.NewEf(err, "could not unmarshal value into type (%T)", result)
	}
	return &value, nil
}

func (r *redisRepo[T]) Drop(c context.Context, key string) error {
	err := r.client.Del(c, key).Err()
	if err != nil {
		return errors.NewEf(err, "could not drop key (%s)", key)
	}
	return nil
}

func NewRedisRepo[T any](redisCli *RedisClient) Repo[T] {
	return &redisRepo[T]{
		RedisClient: redisCli,
	}
}

type RedisClient struct {
	opts   *RedisConnectOptions
	client *redis.Client
}

func (c *RedisClient) Connect(context.Context) error {
	c.client = redis.NewClient(&redis.Options{
		Addr:     c.opts.Addr,
		Password: c.opts.Password,
		Username: c.opts.UserName,
	})
	return nil
}

func (c *RedisClient) Close(context.Context) error {
	err := c.client.Close()
	if err != nil {
		return err
	}
	c.client = nil
	return nil
}

type RedisConnectOptions struct {
	Addr     string
	UserName string
	Password string
}

func NewRedisClient(opts RedisConnectOptions) *RedisClient {
	return &RedisClient{
		opts: &opts,
	}
}
