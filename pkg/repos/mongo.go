package repos

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
	"kloudlite.io/pkg/errors"
)

func NewMongoDatabase(url string, dbName string) (db *mongo.Database, e error) {
	defer errors.HandleErr(&e)

	//structcodec, _ := bsoncodec.NewStructCodec(bsoncodec.JSONFallbackStructTagParser)
	//rb := bson.NewRegistryBuilder()
	//// register struct codec
	//rb.RegisterDefaultEncoder(reflect.Struct, structcodec)
	//
	//client, e := mongo.NewClient(options.Client().SetRegistry(rb.Build()).ApplyURI(url))

	client, e := mongo.NewClient(options.Client().ApplyURI(url))
	errors.AssertNoError(e, fmt.Errorf("could not create mongo client"))
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
						return errors.NewEf(err, "coult not connect to Mongo")
					}
					return db.Client().Ping(ctx, nil)
				},
				OnStop: func(ctx context.Context) error {
					return db.Client().Disconnect(ctx)
				},
			})
		}),
	)
}
