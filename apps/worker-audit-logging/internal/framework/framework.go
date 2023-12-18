package framework

import (
	"github.com/kloudlite/api/apps/worker-audit-logging/internal/app"
	"github.com/kloudlite/api/apps/worker-audit-logging/internal/env"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
	repos "github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
)

type redpandaCfg struct {
	ev *env.Env
}






type eventsDbCfg struct {
	ev *env.Env
}

func (db eventsDbCfg) GetMongoConfig() (url string, dbName string) {
	return db.ev.EventsDbUri, db.ev.EventsDbName
}

var Module fx.Option = fx.Module("framework",
	fx.Provide(func(ev *env.Env) *redpandaCfg {
		return &redpandaCfg{ev: ev}
	}),

	fx.Provide(func(ev *env.Env, logger logging.Logger) (*nats.JetstreamClient, error) {
		name := "audit-worker:jetstream-client"
		nc, err := nats.NewClient(ev.NatsURL, nats.ClientOpts{
			Name:   name,
			Logger: logger,
		})
		if err != nil {
			return nil, err
		}
		return nats.NewJetstreamClient(nc)
	}),

	fx.Provide(func(ev *env.Env) *eventsDbCfg {
		return &eventsDbCfg{ev: ev}
	}),
	repos.NewMongoClientFx[*eventsDbCfg](),
	app.Module,
)
