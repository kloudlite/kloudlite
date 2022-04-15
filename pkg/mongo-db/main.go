package mongo_db

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
	"kloudlite.io/pkg/errors"
)

func NewMongoDatabase(url string, dbName string) (db *mongo.Database, e error) {
	defer errors.HandleErr(&e)
	client, e := mongo.NewClient(options.Client().ApplyURI(url))
	errors.AssertNoError(e, fmt.Errorf("could not create mongo client"))
	return client.Database(dbName), nil
}

type MongoConfig interface {
	GetMongoConfig() (url string, dbName string)
}

func NewFx[T MongoConfig]() fx.Option {
	return fx.Module("db",
		fx.Provide(func(env T) (*mongo.Database, error) {
			return NewMongoDatabase(env.GetMongoConfig())
		}),
		fx.Invoke(func(lifecycle fx.Lifecycle) {
			lifecycle.Append(fx.Hook{
				OnStart: nil,
				OnStop:  nil,
			})
		}),
	)
}
