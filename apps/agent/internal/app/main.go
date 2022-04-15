package app

import (
	"context"
	"encoding/json"
	"strings"

	"go.uber.org/fx"
	"kloudlite.io/apps/agent/internal/domain"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/errors"
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
	callback := func(ctx context.Context, topic string, message messaging.Message) error {
		var msg domain.Message
		err := json.Unmarshal(message, &msg)
		if err != nil {
			return errors.NewEf(err, "could not unmarshal message into (domain.Message)")
		}
		return d.ProcessMessage(ctx, &msg)
	}

	consumer, e := messaging.NewKafkaConsumer(
		messenger, strings.Split(env.Topics, ","), env.GroupId, logger,
		callback,
	)
	if e != nil {
		return nil, e
	}
	return consumer, nil
}

var Module = fx.Module("app",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(fxMsgProducer),
	fx.Provide(fxMsgConsumer),
	domain.Module,
	fx.Invoke(func(lf fx.Lifecycle, consumer messaging.Consumer) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return consumer.Subscribe(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Unsubscribe(ctx)
			},
		})
	}),
	// TModule,
)
