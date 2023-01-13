package framework

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/worker-audit-logging/internal/app"
	"kloudlite.io/apps/worker-audit-logging/internal/env"
	"kloudlite.io/pkg/redpanda"
	repos "kloudlite.io/pkg/repos"
	"strings"
)

type redpandaCfg struct {
	ev *env.Env
}

func (r redpandaCfg) GetSubscriptionTopics() []string {
	return strings.Split(r.ev.KafkaSubscriptionTopics, ",")
}

func (r redpandaCfg) GetConsumerGroupId() string {
	return r.ev.KafkaConsumerGroupId
}

func (r redpandaCfg) GetBrokers() (brokers string) {
	return r.ev.KafkaBrokers
}

func (r redpandaCfg) GetKafkaSASLAuth() *redpanda.KafkaSASLAuth {
	return &redpanda.KafkaSASLAuth{
		SASLMechanism: redpanda.ScramSHA256,
		User:          r.ev.KafkaUsername,
		Password:      r.ev.KafkaPassword,
	}
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
	redpanda.NewClientFx[*redpandaCfg](),
	redpanda.NewConsumerFx[*redpandaCfg](),
	redpanda.NewProducerFx[redpanda.Client](),

	fx.Provide(func(ev *env.Env) *eventsDbCfg {
		return &eventsDbCfg{ev: ev}
	}),
	repos.NewMongoClientFx[*eventsDbCfg](),
	app.Module,
)
