package app

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/build-agent/internal/domain"
	"kloudlite.io/apps/build-agent/internal/env"
	"kloudlite.io/pkg/redpanda"
)

type venv struct {
	ev *env.Env
}

func (fm *venv) GetSubscriptionTopics() []string {
	return []string{fm.ev.KafkaBuildTopics}
}

func (fm *venv) GetConsumerGroupId() string {
	return fm.ev.KafkaConsumerGroup
}

var Module = fx.Module("app",

	redpanda.NewConsumerFx[*venv](),

	fx.Provide(func(ev *env.Env) *venv {
		return &venv{ev}
	}),

	fxInvokeProcessBuilds(),
	domain.Module,
)
