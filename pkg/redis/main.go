package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kloudlite/operator/pkg/errors"
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

func (c *Client) Connect(ctx context.Context) error {
	if err := c.cli.Ping(ctx).Err(); err != nil {
		return errors.NewEf(err, "could not ping to redis host")
	}
	return nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) UpsertUser(ctx context.Context, prefix, username, password string) error {
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
	exists, err := c.userExists(ctx, username)
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return exists, nil
}

func (c *Client) userExists(ctx context.Context, username string) (bool, error) {
	cmd := c.cli.Do(ctx, "ACL", "GETUSER", username)
	_, err := cmd.Result()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) DeleteUser(ctx context.Context, username string) error {
	cmd := c.cli.Do(ctx, "ACL", "DELUSER", username)
	_, err := cmd.Result()
	if err != nil {
		return err
	}
	return nil
}
