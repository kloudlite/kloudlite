package repos

import (
	"context"
	"time"

	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/logging"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

func NewMongoDatabase(ctx context.Context, uri string, dbName string) (db *mongo.Database, e error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errors.NewEf(err, "could not connect to mongodb servers")
	}
	return client.Database(dbName), nil
}

type MongoConfig interface {
	GetMongoConfig() (url string, dbName string)
}

func NewMongoClientFx[T MongoConfig]() fx.Option {
	return fx.Module("db",
		fx.Provide(func(env T) (*mongo.Database, error) {
			url, dbName := env.GetMongoConfig()
			ctx, cf := context.WithTimeout(context.TODO(), 10*time.Second)
			defer cf()
			return NewMongoDatabase(ctx, url, dbName)
		}),

		fx.Invoke(func(db *mongo.Database, logger logging.Logger, lifecycle fx.Lifecycle) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := db.Client().Ping(ctx, nil); err != nil {
						return errors.NewEf(err, "could not ping Mongo")
					}
					logger.Infof("connected to mongodb database: %s", db.Name())
					return nil
				},

				OnStop: func(ctx context.Context) error {
					return db.Client().Disconnect(ctx)
				},
			})
		}),
	)
}
