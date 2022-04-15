package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"kloudlite.io/pkg/errors"
)

type RedisClient struct {
	opts   *RedisConnectOptions
	client *redis.Client
}

func (c *RedisClient) Drop(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return errors.NewEf(err, "could not drop key (%s)", key)
	}
	return nil
}

func (c *RedisClient) SetWithExpiry(
	ctx context.Context,
	key string,
	value []byte,
	duration time.Duration,
) error {
	err := c.client.Set(ctx, key, value, duration).Err()
	if err != nil {
		return errors.NewEf(err, "coult not set key (%s)", key)
	}
	return nil
}

func (c *RedisClient) Connect(ctx context.Context) error {
	rCli := redis.NewClient(&redis.Options{
		Addr:     c.opts.Addr,
		Password: c.opts.Password,
		Username: c.opts.UserName,
	})
	if err := rCli.Ping(ctx).Err(); err != nil {
		return errors.NewEf(err, "could not connect to redis")
	}
	c.client = rCli
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

func (c *RedisClient) Set(ctx context.Context, key string, value []byte) error {
	err := c.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		return errors.NewEf(err, "coult not set key (%s)", key)
	}
	return nil
}

func (c *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, errors.NewEf(err, "could not get key (%s)", key)
	}
	return []byte(result), nil
}

type RedisConnectOptions struct {
	Addr     string
	UserName string
	Password string
}

func NewRedisClient(opts RedisConnectOptions) Client {
	return &RedisClient{
		opts: &opts,
	}
}
