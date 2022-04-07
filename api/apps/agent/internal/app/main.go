package app

import (
	"context"
	"strings"

	"go.uber.org/fx"
	"kloudlite.io/apps/agent/internal/domain"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

var fxNewApp = func(kCli messaging.KafkaClient) {}

type Env struct {
	GroupId string `env:"KAFKA_GROUP_ID", required:"true"`
	Topics  string `env:"KAFKA_TOPICS", required:"true"`
}

func fxMsgProducer(messenger messaging.KafkaClient) (messaging.Producer[domain.Message], error) {
	producer, e := messaging.NewKafkaProducer[domain.Message](messenger)
	if e != nil {
		return nil, e
	}
	return producer, nil
}

func fxMsgConsumer(messenger messaging.KafkaClient, env *Env, logger logger.Logger, d domain.Domain) (messaging.Consumer[domain.Message], error) {
	consumer, e := messaging.NewKafkaConsumer[domain.Message](
		messenger, strings.Split(env.Topics, ","), env.GroupId, logger,
		func(topic string, msg domain.Message) error {
			return nil
		},
	)
	if e != nil {
		return nil, e
	}
	return consumer, nil
}

var Module = fx.Module("app",
	fx.Provide(fxNewApp),
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(fxMsgProducer),
	fx.Provide(fxMsgConsumer),
	domain.Module,
	fx.Invoke(func(lf fx.Lifecycle, consumer messaging.Consumer[domain.Message]) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return consumer.Subscribe()
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Unsubscribe()
			},
		})
	}),
)
