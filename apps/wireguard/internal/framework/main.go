package framework

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"kloudlite.io/pkg/config"
	gql_server "kloudlite.io/pkg/gql-server"
	"kloudlite.io/pkg/logger"
	mongo_db "kloudlite.io/pkg/mongo-db"

	"go.uber.org/fx"
	"kloudlite.io/apps/wireguard/internal/app"
	"net/http"
)

type Env struct {
	MongoUri    string `env:"MONGO_URI", required:"true"`
	MongoDbName string `env:"MONGO_DB_NAME", required:"true"`
	Port        uint32 `env:"PORT", required:"true"`
}

var Module = fx.Module("framework",
	// Setup Logger
	fx.Provide(logger.NewLogger),
	// Load Env
	fx.Provide(func() (*Env, error) {
		var envC Env
		err := config.LoadEnv(&envC)
		return &envC, err
	}),
	// Create DB Client
	fx.Provide(func(env *Env) (*mongo.Database, error) {
		return mongo_db.NewMongoDatabase(env.MongoUri, env.MongoDbName)
	}),
	// Connect DB Client
	fx.Invoke(func(lifecycle fx.Lifecycle, db *mongo.Database) {
		lifecycle.Append(fx.Hook{
			OnStart: func(context.Context) error {
				return db.Client().Connect(context.Background())
			},
		})
	}),
	// Load App Module
	app.Module,
	// start http server
	fx.Invoke(func(lf fx.Lifecycle, env *Env, logger logger.Logger, gqlHandler http.Handler) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return gql_server.StartGQLServer(ctx, env.Port, gqlHandler, logger)
			},
			OnStop: func(context.Context) error {
				return nil
			},
		})
	}),
)
