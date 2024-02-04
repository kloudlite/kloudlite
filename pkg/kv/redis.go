package kv

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kloudlite/api/pkg/errors"
)

type RedisClient struct {
	opts       *redis.Options
	client     *redis.Client
	basePrefix string
}

func (c *RedisClient) Connect(ctx context.Context) error {
	rCli := redis.NewClient(c.opts)
	if err := rCli.Ping(ctx).Err(); err != nil {
		return errors.NewEf(err, "could not connect to redis")
	}
	c.client = rCli
	return nil
}

func (c *RedisClient) Disconnect(context.Context) error {
	err := c.client.Close()
	if err != nil {
		return errors.NewE(err)
	}
	c.client = nil
	return nil
}

func (c *RedisClient) getKey(key string) string {
	if c.basePrefix != "" {
		return fmt.Sprintf("%s:%s", c.basePrefix, key)
	}
	return key
}

func (c *RedisClient) Set(ctx context.Context, key string, value []byte) error {
	k := c.getKey(key)
	err := c.client.Set(ctx, k, value, 0).Err()
	if err != nil {
		return errors.NewEf(err, "could not set key (%s)", k)
	}
	return nil
}

func (c *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	if c.client == nil {
		return nil, errors.Newf("redis client is not connected")
	}
	status := c.client.Get(ctx, c.getKey(key))
	err := status.Err()
	if err != nil {
		return nil, errors.NewE(err)
	}
	return []byte(status.Val()), nil
}

func (c *RedisClient) Drop(ctx context.Context, key string) error {
	k := c.getKey(key)
	err := c.client.Del(ctx, k).Err()
	if err != nil {
		return errors.NewEf(err, "could not drop key=%s", k)
	}

	return nil
}

func (c *RedisClient) SetWithExpiry(
	ctx context.Context,
	key string,
	value []byte,
	duration time.Duration,
) error {
	if c.client == nil {
		return errors.Newf("redis client is not connected")
	}
	k := c.getKey(key)
	err := c.client.Set(ctx, k, value, duration).Err()
	if err != nil {
		return errors.NewEf(err, "could not set value for key=%s", k)
	}
	return nil
}

func NewRedisClient(hosts, username, password, basePrefix string) Client {
	return &RedisClient{
		opts: &redis.Options{
			Addr:     hosts,
			Username: username,
			Password: password,
		},
		basePrefix: basePrefix,
	}
}

type TypedRedisClient[T Client] struct {
	opts   *redis.Options
	client T
}

type RedisConfig interface {
	RedisOptions() (hosts, username, password, basePrefix string)
}

type RedisConfigTyped[T Client] interface {
	RedisOptions() (hosts, username, password, basePrefix string)
}

func FxLifeCycle[T Client]() fx.Option {
	return fx.Module(
		"redis-fx-lifecycle",
		fx.Invoke(
			func(c T, lf fx.Lifecycle) {
				lf.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							return c.Connect(ctx)
						},
						OnStop: func(ctx context.Context) error {
							return c.Disconnect(ctx)
						},
					},
				)
			},
		),
	)
}

func NewRedisFx[T RedisConfig]() fx.Option {
	return fx.Module(
		"redis",
		fx.Provide(
			func(env T) Client {
				options, username, password, basePrefix := env.RedisOptions()
				return NewRedisClient(options, username, password, basePrefix)
			},
		),
		fx.Invoke(
			func(lf fx.Lifecycle, r Client) {
				lf.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							return r.Connect(ctx)
						},
						OnStop: func(ctx context.Context) error {
							return r.Disconnect(ctx)
						},
					},
				)
			},
		),
	)
}
