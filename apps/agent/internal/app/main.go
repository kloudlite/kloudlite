package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/fx"
	"kloudlite.io/apps/agent/internal/domain"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

type Env struct {
	KafkaGroupId string `env:"KAFKA_GROUP_ID" required:"true"`
	KafkaTopic   string `env:"KAFKA_TOPIC" required:"true"`
}

type M map[string]interface{}

func fxMsgProducer(messenger messaging.KafkaClient) (messaging.Producer[domain.MessageReply], error) {
	producer, e := messaging.NewKafkaProducer[domain.MessageReply](messenger)
	if e != nil {
		return nil, e
	}
	return producer, nil
}

func fxMsgConsumer(messenger messaging.KafkaClient, env *Env, logger logger.Logger, d domain.Domain) (messaging.Consumer, error) {
	callback := func(ctx context.Context, topic string, message messaging.Message) error {
		fmt.Println("#$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
		var msg domain.Message
		err := json.Unmarshal(message, &msg)
		if err != nil {
			return errors.NewEf(err, "could not unmarshal message into (domain.Message)")
		}
		return d.ProcessMessage(ctx, &msg)
	}

	consumer, e := messaging.NewKafkaConsumer(
		messenger,
		strings.Split(env.KafkaTopic, ","),
		env.KafkaGroupId,
		logger,
		callback,
	)
	if e != nil {
		return nil, e
	}
	return consumer, nil
}

var Module = fx.Module("app",
	config.EnvFx[Env](),
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
	TModule,

	// DEV module
	// fx.Invoke(func(lf fx.Lifecycle, env *Env, producer messaging.Producer[domain.Message]) {
	// 	lf.Append(fx.Hook{
	// 		OnStart: func(ctx context.Context) error {
	// 			err := producer.Connect(ctx)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			err = producer.SendMessage(env.KafkaTopic, "hello-world", domain.Message{
	// 				ResourceType: common.ResourceProject,
	// 				Namespace:    "sample",
	// 				Spec: domain.Project{
	// 					Name:        "sample",
	// 					DisplayName: "this is just a sample project",
	// 				},
	// 			})
	// 			fmt.Println("MESSAGE dumped into kafka")
	// 			return err
	// 		},
	// 	})
	// }),
)
