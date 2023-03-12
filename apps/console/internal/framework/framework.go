package framework

import (
	"go.uber.org/fx"
	app "kloudlite.io/apps/console/internal/app"
	"kloudlite.io/apps/console/internal/env"
	"kloudlite.io/pkg/cache"
	httpServer "kloudlite.io/pkg/http-server"
	mongoDb "kloudlite.io/pkg/repos"
)

type fm struct {
	ev *env.Env
}

func (fm *fm) GetMongoConfig() (url string, dbName string) {
	return fm.ev.ConsoleDBUri, fm.ev.ConsoleDBName
}

func (fm *fm) RedisOptions() (hosts, username, password, basePrefix string) {
	return fm.ev.AuthRedisHosts, fm.ev.AuthRedisUserName, fm.ev.AuthRedisPassword, fm.ev.AuthRedisPrefix
}

func (fm *fm) GetHttpPort() uint16 {
	return fm.ev.Port
}

func (fm *fm) GetHttpCors() string {
	return ""
}

var Module = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *fm {
		return &fm{ev}
	}),
	mongoDb.NewMongoClientFx[*fm](),
	fx.Provide(func(ev *env.Env) app.AuthCacheClient {
		return cache.NewRedisClient(ev.AuthRedisHosts, ev.AuthRedisUserName, ev.AuthRedisPassword, ev.AuthRedisPrefix)
	}),
	cache.FxLifeCycle[app.AuthCacheClient](),
	app.Module,
	httpServer.NewHttpServerFx[*fm](),
)
