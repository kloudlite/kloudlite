package mongo

import (
	"context"
	"fmt"
	"github.com/kloudlite/operator/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Client struct {
	conn        *mongo.Client
	isConnected bool
}

var (
	ErrNotConnected = fmt.Errorf("is not connected to db yet, call Connect() method")
)

func NewClient(uri string) (*Client, error) {
	cli, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errors.NewEf(err, "could not create mongodb client")
	}
	return &Client{conn: cli}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	if err := c.conn.Connect(ctx); err != nil {
		return errors.NewEf(err, "could not connect to specified mongodb service")
	}
	if err := c.conn.Ping(ctx, &readpref.ReadPref{}); err != nil {
		return errors.NewEf(err, "could not ping mongodb")
	}
	c.isConnected = true
	return nil
}

func (c *Client) Close() error {
	return c.conn.Disconnect(context.TODO())
}

func (c *Client) UpsertUser(ctx context.Context, dbName string, userName string, password string) error {
	if !c.isConnected {
		return ErrNotConnected
	}

	if v, _ := c.userExists(ctx, "", userName); v {
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
	if !c.isConnected {
		return false, ErrNotConnected
	}

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
