package cache

import (
	"context"
	"go.uber.org/fx"
	"time"

	"github.com/go-redis/redis/v8"
	"kloudlite.io/pkg/errors"
)

type RedisClient struct {
	opts   *redis.Options
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
	rCli := redis.NewClient(c.opts)
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
	status := c.client.Get(ctx, key)
	err := status.Err()
	if err != nil {
		return nil, err
	}
	return []byte(status.Val()), nil
}

func NewRedisClient(hosts, username, password string) Client {
	return &RedisClient{
		opts: &redis.Options{
			Addr:     hosts,
			Username: username,
			Password: password,
		},
	}
}

type RedisConfig interface {
	RedisOptions() (hosts, username, password string)
}

func NewRedisFx[T RedisConfig]() fx.Option {
	return fx.Module("redis",
		fx.Provide(func(env T) Client {
			return NewRedisClient(env.RedisOptions())
		}),
		fx.Invoke(func(lf fx.Lifecycle, r Client) {
			lf.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return r.Connect(ctx)
				},
				OnStop: func(ctx context.Context) error {
					return r.Close(ctx)
				},
			})
		}),
	)
}
