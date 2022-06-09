package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/dns/internal/app"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/dns"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	DNSPort       uint16 `env:"DNS_PORT" required:"true"`
	DBName        string `env:"MONGO_DB_NAME"`
	DBUrl         string `env:"MONGO_URI"`
	RedisHosts    string `env:"REDIS_HOSTS"`
	RedisUsername string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	HttpPort      uint16 `env:"PORT"`
	HttpCors      string `env:"ORIGINS"`
	GrpcPort      uint16 `env:"GRPC_PORT"`
}

func (e *Env) GetDNSPort() uint16 {
	return e.DNSPort
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
	config.EnvFx[Env](),
	fx.Provide(logger.NewLogger),
	cache.NewRedisFx[*Env](),
	repos.NewMongoClientFx[*Env](),
	httpServer.NewHttpServerFx[*Env](),
	dns.Fx[*Env](),
	app.Module,
)
