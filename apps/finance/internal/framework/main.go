package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/finance/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	DBName        string `env:"DB_NAME"`
	DBUrl         string `env:"DB_URL"`
	RedisHosts    string `env:"REDIS_HOSTS"`
	RedisUsername string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	HttpPort      uint16 `env:"HTTP_PORT"`
	HttpCors      string `env:"ORIGINS"`
}

func (e *Env) GetMongoConfig() (url string, dbName string) {
	return e.DBUrl, e.DBName
}

func (e *Env) RedisOptions() (hosts, username, password string) {
	return e.RedisHosts, e.RedisUsername, e.RedisPassword
}

func (e *Env) GetHttpPort() uint16 {
	return e.HttpPort
}

func (e *Env) GetHttpCors() string {
	return e.HttpCors
}

var Module = fx.Module("framework",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(logger.NewLogger),
	repos.NewMongoClientFx[*Env](),
	cache.NewRedisFx[*Env](),
	httpServer.NewHttpServerFx[*Env](),
	app.Module,
)
