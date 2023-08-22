package repos

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/fx"
	"kloudlite.io/pkg/errors"
)

func NewMongoDatabase(url string, dbName string) (db *mongo.Database, e error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		return nil, errors.NewEf(err, "could not create mongo client")
	}
	return client.Database(dbName), nil
}

type MongoConfig interface {
	GetMongoConfig() (url string, dbName string)
}

func NewMongoClientFx[T MongoConfig]() fx.Option {
	return fx.Module("db",
		fx.Provide(func(env T) (*mongo.Database, error) {
			return NewMongoDatabase(env.GetMongoConfig())
		}),

		fx.Invoke(func(db *mongo.Database, lifecycle fx.Lifecycle) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					err := db.Client().Connect(ctx)
					if err != nil {
						return errors.NewEf(err, "could not connect to Mongo")
					}
					if err := db.Client().Ping(ctx, nil); err != nil {
						return errors.NewEf(err, "could not ping Mongo")
					}

					if err = db.Client().Ping(ctx, &readpref.ReadPref{}); err != nil {
						return errors.NewEf(err, "failed to ping mongo")
					}
					return nil
				},

				OnStop: func(ctx context.Context) error {
					return db.Client().Disconnect(ctx)
				},
			})
		}),
	)
}
