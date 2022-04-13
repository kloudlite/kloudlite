package cache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"
)

type redisRepo[T any] struct {
	*RedisClient
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

func NewRedisRepo[T any](redisCli *RedisClient) Repo[T] {
	return &redisRepo[T]{
		RedisClient: redisCli,
	}
}

type RedisClient struct {
	opts struct {
		Addr     string
		Username string
		Password string
	}
	client *redis.Client
}

func (c *RedisClient) Connect(context.Context) error {
	c.client = redis.NewClient(&redis.Options{
		Addr:     c.opts.Addr,
		Password: c.opts.Password,
		Username: c.opts.Username,
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

func NewRedisClient(addr string, password string, userName string) *RedisClient {
	return &RedisClient{
		opts: struct {
			Addr     string
			Username string
			Password string
		}{Addr: addr, Username: userName, Password: password},
	}
}
