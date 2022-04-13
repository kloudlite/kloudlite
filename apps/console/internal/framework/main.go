package framework

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/app"
	"kloudlite.io/pkg/config"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
	mongo_db "kloudlite.io/pkg/mongo-db"
)

type Env struct {
	MongoUri     string `env:"MONGO_URI" required:"true"`
	MongoDbName  string `env:"MONGO_DB_NAME" required:"true"`
	KafkaBrokers string `env:"KAFKA_BOOTSTRAP_SERVERS" required:"true"`
	Port         uint16 `env:"PORT" required:"true"`
	IsDev        bool   `env:"DEV" default:"false"`
	CorsOrigins  string `env:"ORIGINS"`
}

var Module = fx.Module("framework",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(logger.NewLogger),

	// Create DB Client
	fx.Provide(func(env *Env) (*mongo.Database, error) {
		return mongo_db.NewMongoDatabase(env.MongoUri, env.MongoDbName)
	}),

	fx.Provide(http.NewServeMux),

	fx.Provide(func(e *Env) messaging.KafkaClient {
		return messaging.NewKafkaClient(e.KafkaBrokers)
	}),

	// Load App Module
	app.Module,

	// Connect DB Client
	fx.Invoke(func(lf fx.Lifecycle, db *mongo.Database) {
		lf.Append(fx.Hook{
			OnStart: func(pCtx context.Context) error {
				ctx, cancelFn := context.WithTimeout(pCtx, time.Second*2)
				defer cancelFn()
				e := db.Client().Connect(ctx)
				if e != nil {
					return e
				}
				return db.Client().Ping(ctx, &readpref.ReadPref{})
			},
			OnStop: func(ctx context.Context) error {
				return db.Client().Disconnect(ctx)
			},
		})
	}),

	// start http server
	fx.Invoke(func(lf fx.Lifecycle, env *Env, logger logger.Logger, server *http.ServeMux) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				corsOpt := cors.Options{
					AllowedOrigins:   strings.Split(env.CorsOrigins, ","),
					AllowCredentials: true,
					AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
				}
				return httpServer.Start(ctx, env.Port, server, corsOpt, logger)
			},
		})
	}),
)
