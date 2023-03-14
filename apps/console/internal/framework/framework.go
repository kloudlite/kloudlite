package framework

import (
	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
	app "kloudlite.io/apps/console/internal/app"
	"kloudlite.io/apps/console/internal/env"
	"kloudlite.io/pkg/cache"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/k8s"
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
	return "*"
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

	fx.Provide(func(restCfg *rest.Config) (*kubectl.YAMLClient, error) {
		return kubectl.NewYAMLClient(restCfg)
	}),

	fx.Provide(func(restCfg *rest.Config) (k8s.ExtendedK8sClient, error) {
		return k8s.NewExtendedK8sClient(restCfg)
	}),

	app.Module,
	httpServer.NewHttpServerFx[*fm](),
)
