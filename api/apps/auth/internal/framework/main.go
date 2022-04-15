package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/auth/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	MongoUri      string `env:"MONGO_URI" required:"true"`
	RedisHosts    string `env:"REDIS_HOSTS" required:"true"`
	RedisUserName string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	MongoDbName   string `env:"MONGO_DB_NAME" required:"true"`
	Port          uint16 `env:"PORT" required:"true"`
	CorsOrigins   string `env:"ORIGINS"`
}

func (e *Env) GetHttpPort() uint16 {
	return e.Port
}

func (e *Env) GetHttpCors() string {
	return e.CorsOrigins
}

func (e *Env) RedisOptions() (hosts, username, password string) {
	return e.RedisHosts, e.RedisUserName, e.RedisPassword
}

func (e *Env) GetMongoConfig() (url string, dbName string) {
	return e.MongoUri, e.MongoDbName
}

var Module = fx.Module("framework",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(logger.NewLogger),
	repos.NewMongoClientFx[*Env](),
	cache.NewRedisFx[*Env](),
	httpServer.NewHttpServerFx[*Env](),
	app.Module,
)
