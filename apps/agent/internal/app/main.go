package app

import (
	"strings"

	"go.uber.org/fx"
	"kloudlite.io/apps/agent/internal/domain"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

type Env struct {
	GroupId string `env:"KAFKA_GROUP_ID" required:"true"`
	Topics  string `env:"KAFKA_TOPICS" required:"true"`
}

type M map[string]interface{}

func fxMsgProducer(messenger messaging.KafkaClient) (messaging.Producer[domain.Message], error) {
	producer, e := messaging.NewKafkaProducer[domain.Message](messenger)
	if e != nil {
		return nil, e
	}
	return producer, nil
}

func fxMsgConsumer(messenger messaging.KafkaClient, env *Env, logger logger.Logger, d domain.Domain) (messaging.Consumer, error) {
	consumer, e := messaging.NewKafkaConsumer(
		messenger, strings.Split(env.Topics, ","), env.GroupId, logger,
		func(topic string, msg domain.Message) error {
			return d.ProcessMessage(&msg)
		},
	)
	if e != nil {
		return nil, e
	}
	return consumer, nil
}

var Module = fx.Module("app",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(fxMsgProducer),
	// fx.Provide(fxMsgConsumer),
	domain.Module,
	// fx.Invoke(func(lf fx.Lifecycle, consumer messaging.Consumer) {
	// 	lf.Append(fx.Hook{
	// 		OnStart: func(ctx context.Context) error {
	// 			return consumer.Subscribe()
	// 			// return nil
	// 		},
	// 		OnStop: func(ctx context.Context) error {
	// 			// return consumer.Unsubscribe()
	// 			return nil
	// 		},
	// 	})
	// }),
	TModule,
)
