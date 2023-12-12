package app

import (
	"github.com/kloudlite/api/apps/build-agent/internal/domain"
	"github.com/kloudlite/api/apps/build-agent/internal/env"
	"github.com/kloudlite/api/pkg/redpanda"
	"go.uber.org/fx"
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
