package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/dns/internal/app"

	// "kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/dns"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	DNSPort       uint16 `env:"DNS_PORT" required:"true"`
	MongoUri      string `env:"MONGO_URI" required:"true"`
	RedisHosts    string `env:"REDIS_HOSTS" required:"true"`
	RedisUserName string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisPrefix   string `env:"REDIS_PREFIX"`
	MongoDbName   string `env:"MONGO_DB_NAME" required:"true"`
	Port          uint16 `env:"PORT" required:"true"`
	IsDev         bool   `env:"DEV" default:"false" required:"true"`
	CorsOrigins   string `env:"ORIGINS"`
}

func (e *Env) GetDNSPort() uint16 {
	return e.DNSPort
}

func (e *Env) GetMongoConfig() (url string, dbName string) {
	return e.MongoUri, e.MongoDbName
}

func (e *Env) RedisOptions() (hosts, username, password, prefix string) {
	return e.RedisHosts, e.RedisUserName, e.RedisPassword, e.RedisPrefix
}

func (e *Env) GetHttpPort() uint16 {
	return e.Port
}

func (e *Env) GetHttpCors() string {
	return e.CorsOrigins
}

var Module = fx.Module(
	"framework",
	config.EnvFx[Env](),
	repos.NewMongoClientFx[*Env](),
	cache.NewRedisFx[*Env](),
	httpServer.NewHttpServerFx[*Env](),
	dns.Fx[*Env](),
	app.Module,
)
