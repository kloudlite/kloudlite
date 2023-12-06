package mongo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kloudlite/operator/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	conn        *mongo.Client
	isConnected bool
}

var ErrNotConnected = fmt.Errorf("is not connected to db yet, call Connect() method")

func newClient(ctx context.Context, uri string) (*Client, error) {
	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errors.NewEf(err, "could not create mongodb client")
	}
	return &Client{conn: cli}, nil
}

func NewClient(ctx context.Context, uri string) (*Client, error) {
	return newClient(ctx, uri)
}

func (c *Client) ValidateAuthenticatedURI(ctx context.Context, uri string) error {
	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return errors.NewEf(err, "could not connect to specified mongodb service")
	}
	if err := cli.Ping(ctx, nil); err != nil {
		return errors.NewEf(err, "could not ping mongodb")
	}

	return nil
}

func (c *Client) Close() error {
	return c.conn.Disconnect(context.TODO())
}

func (c *Client) UpsertUser(ctx context.Context, dbName string, userName string, password string) error {
	if v, _ := c.userExists(ctx, dbName, userName); v {
		return nil
	}

	db := c.conn.Database(dbName)

	var user bson.M
	err := db.RunCommand(
		ctx, bson.D{
			{Key: "createUser", Value: userName},
			{Key: "pwd", Value: password},
			{
				Key: "roles", Value: []bson.M{
					{"role": "dbAdmin", "db": dbName},
					{"role": "readWrite", "db": dbName},
				},
			},
		},
	).Decode(&user)
	if err != nil {
		return errors.NewEf(err, "could not create user")
	}
	return nil
}

func (c *Client) UserExists(ctx context.Context, dbName string, userName string) (bool, error) {
	return c.userExists(ctx, dbName, userName)
}

func (c *Client) UpdateUserPassword(ctx context.Context, dbName string, userName string, password string) error {
	if !c.isConnected {
		return ErrNotConnected
	}

	exists, err := c.userExists(ctx, dbName, userName)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("user %q does not exist", userName)
	}

	db := c.conn.Database(dbName)
	return db.RunCommand(
		ctx, bson.D{
			{Key: "updateUser", Value: userName},
			{Key: "pwd", Value: password},
		},
	).Err()
}

func (c *Client) DeleteUser(ctx context.Context, dbName string, username string) error {
	if exists, _ := c.userExists(ctx, "", username); !exists {
		return nil
	}
	db := c.conn.Database(dbName)
	return db.RunCommand(
		ctx, bson.D{
			{Key: "dropUser", Value: username},
		},
	).Err()
}

func (c *Client) userExists(ctx context.Context, dbName string, userName string) (bool, error) {
	db := c.conn.Database(dbName)
	sr := db.RunCommand(
		ctx, bson.D{
			{Key: "usersInfo", Value: userName},
		},
	)

	var usersInfo struct {
		Users []interface{} `json:"users" bson:"users"`
	}

	if err := sr.Decode(&usersInfo); err != nil {
		return false, errors.NewEf(err, "could not decode usersInfo")
	}

	return len(usersInfo.Users) > 0, nil
}

func ConnectAndPing(ctx context.Context, authenticatedUri string) error {
	cli, err := newClient(ctx, authenticatedUri)
	defer func() {
		if err := cli.Close(); err != nil {
			fmt.Println("could not close client")
		}
	}()
	if err != nil {
		return err
	}

	if err := cli.conn.Ping(ctx, nil); err != nil {
		return errors.NewEf(err, "could not ping mongodb")
	}

	return nil
}

func FailsWithAuthError(err error) bool {
	if err != nil {
		return strings.Contains(err.Error(), "(AuthenticationFailed)")
	}
	return false
}
