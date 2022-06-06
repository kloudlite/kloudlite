package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"operators.kloudlite.io/lib/errors"
)

type Client struct {
	cli *redis.Client
}

func NewClient(hosts, username, password string) (*Client, error) {
	rCli := redis.NewClient(
		&redis.Options{
			Addr:     hosts,
			Username: username,
			Password: password,
		},
	)
	if rCli == nil {
		return nil, errors.Newf("could not build redis client")
	}
	return &Client{cli: rCli}, nil
}

func (c *Client) ping(ctx context.Context) error {
	if err := c.cli.Ping(ctx).Err(); err != nil {
		return errors.NewEf(err, "could not ping to redis host")
	}
	return nil
}

func (c *Client) UpsertUser(ctx context.Context, prefix, username, password string) error {
	if err := c.ping(ctx); err != nil {
		return err
	}
	if err := c.cli.Do(
		ctx,
		"ACL", "SETUSER", username, "on",
		fmt.Sprintf("~%s:*", prefix),
		"+@all", "-@dangerous", "+info", "resetpass", fmt.Sprintf(">%s", password),
	).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Client) UserExists(ctx context.Context, username string) (bool, error) {
	return c.userExists(ctx, username)
}

func (c *Client) userExists(ctx context.Context, username string) (bool, error) {
	if err := c.ping(ctx); err != nil {
		return false, err
	}
	cmd := c.cli.Do(ctx, "ACL", "GETUSER", username)
	_, err := cmd.Result()
	if err != nil {
		return false, err
	}
	return true, nil
}
