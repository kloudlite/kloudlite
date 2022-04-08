package app

import (
	"context"
	"strings"

	"go.uber.org/fx"
	"kloudlite.io/apps/agent/internal/domain"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/shared"
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
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(fxMsgProducer),
	fx.Provide(fxMsgConsumer),
	domain.Module,
	fx.Invoke(func(lf fx.Lifecycle, consumer messaging.Consumer) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return nil
				// return consumer.Subscribe()
			},
			OnStop: func(ctx context.Context) error {
				// return consumer.Unsubscribe()
				return nil
			},
		})
	}),

	fx.Invoke(func(lf fx.Lifecycle, d domain.Domain) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {

				// msg := domain.Message{
				// 	ResourceType: shared.RESOURCE_PROJECT,
				// 	Namespace:    "hotspot",
				// 	Spec: domain.Project{
				// 		Name:        "sample-xyz",
				// 		DisplayName: "this is not just a project",
				// 		Logo:        "i have no logo",
				// 	},
				// }

				// msg := domain.Message{
				// 	ResourceType: shared.RESOURCE_MANAGED_SERVICE,
				// 	Namespace:    "hotspot",
				// 	Spec: domain.ManagedSvc{
				// 		Name:         "sample-xyz",
				// 		Namespace:    "hotspot",
				// 		TemplateName: "msvc_mongo",
				// 		Version:      1,
				// 		Values: map[string]interface{}{
				// 			"hi": "asdfa",
				// 		},
				// 		LastApplied: M{"hello": "world", "something": map[string]interface{}{
				// 			"one": 2,
				// 			"two": 2,
				// 		}},
				// 	},
				// }

				// msg := domain.Message{
				// 	ResourceType: shared.RESOURCE_APP,
				// 	Namespace:    "hotspot",
				// 	Spec: domain.App{
				// 		Name:      "sample",
				// 		Namespace: "hotspot",
				// 		Services: []domain.AppSvc{
				// 			domain.AppSvc{
				// 				Port:       21323,
				// 				TargetPort: 21345,
				// 				Type:       "tcp",
				// 			},
				// 		},
				// 		Containers: []domain.AppContainer{
				// 			domain.AppContainer{
				// 				Name:            "sample",
				// 				Image:           "nginx",
				// 				ImagePullPolicy: "Always",
				// 				Command:         []string{"hello", "world"},
				// 				ResourceCpu:     domain.ContainerResource{Min: "100", Max: "200"},
				// 				ResourceMemory:  domain.ContainerResource{Min: "200", Max: "300"},
				// 				Env: []domain.ContainerEnv{
				// 					domain.ContainerEnv{
				// 						Key:   "hello",
				// 						Value: "world",
				// 					},
				// 				},
				// 			},
				// 		},
				// 	},
				// }

				// msg := domain.Message{
				// 	ResourceType: shared.RESOURCE_MANAGED_RESOURCE,
				// 	Namespace:    "hotspot",
				// 	Spec: domain.ManagedRes{
				// 		Name:       "sample-mres",
				// 		Type:       "db",
				// 		Namespace:  "hotspot",
				// 		ManagedSvc: "sample1234",
				// 		Values: map[string]interface{}{
				// 			"hello":  "world",
				// 			"sample": "hello",
				// 		},
				// 	},
				// }

				// msg := domain.Message{
				// 	ResourceType: shared.RESOURCE_CONFIG,
				// 	Namespace:    "hotspot",
				// 	Spec: domain.Config{
				// 		Name:      "hi-config",
				// 		Namespace: "hotspot",
				// 		Data: map[string]interface{}{
				// 			"hi":  "hello there",
				// 			"one": 2,
				// 		},
				// 	},
				// }

				msg := domain.Message{
					ResourceType: shared.RESOURCE_SECRET,
					Namespace:    "hotspot",
					Spec: domain.Secret{
						Name:      "hi-config",
						Namespace: "hotspot",
						Data: map[string]interface{}{
							"hi":  "hello there",
							"one": 2,
						},
					},
				}

				return d.ProcessMessage(&msg)
			},
		})
	}),
)
