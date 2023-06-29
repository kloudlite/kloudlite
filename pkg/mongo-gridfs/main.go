package mongogridfs

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

type MongoConfig interface {
	GetMongoConfig() (url string, dbName string)
}

func NewMongoGridFsClientFx[T MongoConfig]() fx.Option {
	return fx.Module("mongodb-gridfs",
		fx.Provide(
			func(env T) (*gridfs.Bucket, GridFs, error) {

				ctx, cancel := context.WithTimeout(
					context.Background(),
					10*time.Second,
				)
				defer cancel()

				url, dbName := env.GetMongoConfig()

				client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
				if err != nil {
					return nil, nil, err
				}

				db := client.Database(dbName)
				bucket, err := gridfs.NewBucket(db)
				if err != nil {
					return nil, nil, err
				}

				gridfs := &gfs{
					bucket: bucket,
				}

				return bucket, gridfs, nil
			},
		),
	)
}
