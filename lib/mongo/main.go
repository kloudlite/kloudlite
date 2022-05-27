package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"operators.kloudlite.io/lib/errors"
)

type Client struct {
	conn *mongo.Client
}

func NewClient(uri string) (*Client, error) {
	cli, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errors.NewEf(err, "could not create mongodb client")
	}
	return &Client{conn: cli}, nil
}

func (c *Client) connect(ctx context.Context) error {
	if err := c.conn.Ping(ctx, &readpref.ReadPref{}); err != nil {
		if err := c.conn.Connect(ctx); err != nil {
			return errors.NewEf(err, "could not connect to specified mongodb service")
		}
	}
	return nil
}

func (c *Client) UpsertUser(ctx context.Context, dbName string, userName string, password string) error {
	if err := c.connect(ctx); err != nil {
		return err
	}

	defer c.conn.Disconnect(ctx)

	if v, _ := c.userExists(ctx, userName); v {
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

func (c *Client) UserExists(ctx context.Context, userName string) (bool, error) {
	return c.userExists(ctx, userName)
}

func (c *Client) userExists(ctx context.Context, userName string) (bool, error) {
	if err := c.connect(ctx); err != nil {
		return false, err
	}

	db := c.conn.Database(userName)
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
