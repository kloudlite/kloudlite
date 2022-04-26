package framework

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/ci/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	fiberapp "kloudlite.io/pkg/fiber-app"
	rpc "kloudlite.io/pkg/grpc"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	DBName        string `env:"MONGO_DB_NAME"`
	DBUrl         string `env:"MONGO_URI"`
	RedisHost     string `env:"REDIS_HOST"`
	RedisUsername string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	HttpPort      uint16 `env:"PORT" required:"true"`
	HttpCors      string `env:"ORIGINS"`
	GrpcPort      uint16 `env:"GRPC_PORT"`
}

type GrpcAuthConfig struct {
	InfraGrpcHost string `env:"AUTH_HOST" required:"true"`
	InfraGrpcPort string `env:"AUTH_PORT" required:"true"`
}

func (e *GrpcAuthConfig) GetGCPServerURL() string {
	return e.InfraGrpcHost + ":" + e.InfraGrpcPort
}

func (e *Env) GetGRPCPort() uint16 {
	return e.GrpcPort
}

func (e *Env) RedisOptions() (hosts, username, password string) {
	return e.RedisUsername, e.RedisPassword, e.RedisHost
}

func (e *Env) GetMongoConfig() (url string, dbName string) {
	return e.DBUrl, e.DBName
}

var Module = fx.Module("framework",
	fx.Provide(logger.NewLogger),
	config.EnvFx[Env](),
	config.EnvFx[GrpcAuthConfig](),
	repos.NewMongoClientFx[*Env](),
	cache.NewRedisFx[*Env](),
	fx.Provide(fiberapp.NewFiberApp),
	rpc.NewGrpcServerFx[*Env](),
	rpc.NewGrpcClientFx[*GrpcAuthConfig, app.AuthClientConnection](),
	fx.Invoke(func(app *fiber.App, env *Env, lifecycle fx.Lifecycle) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go app.Listen(fmt.Sprintf(":%v", env.HttpPort))
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return nil
			},
		})
	}),
	app.Module,
)
