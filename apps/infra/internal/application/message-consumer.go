package application

import (
	"context"
	"fmt"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

func fxConsumer(env *InfraEnv, mc messaging.KafkaClient, d domain.Domain, logger logger.Logger) (messaging.Consumer, error) {
	fmt.Println("fxConsumer", env.KafkaInfraTopic, env.KafkaGroupId)
	consumer, err := messaging.NewKafkaConsumer(
		mc,
		[]string{env.KafkaInfraTopic},
		env.KafkaGroupId,
		logger,
		func(context context.Context, topic string, action messaging.Message) error {
			var _d struct {
				Type string
			}
			action.Unmarshal(&_d)
			fmt.Println(_d.Type)
			switch _d.Type {
			case "create-cluster":
				var m struct {
					Type    string
					Payload domain.SetupClusterAction
				}
				action.Unmarshal(&m)
				return d.CreateCluster(context, m.Payload)
				break
			case "update-cluster":
				var m struct {
					Type    string
					Payload domain.UpdateClusterAction
				}
				action.Unmarshal(&m)
				logger.Info("message", "type", m)
				return d.UpdateCluster(context, m.Payload)
				break
			case "delete-cluster":
				var m struct {
					Type    string
					Payload domain.DeleteClusterAction
				}
				action.Unmarshal(&m)
				logger.Info("message", "type", m)
				return d.DeleteCluster(context, m.Payload)
				break
			case "add-peer":
				var m struct {
					Type    string
					Payload domain.AddPeerAction
				}
				action.Unmarshal(&m)
				logger.Info("message", "type", m)
				return d.AddPeerToCluster(context, m.Payload)
				break
			case "delete-peer":
				var m struct {
					Type    string
					Payload domain.DeletePeerAction
				}
				action.Unmarshal(&m)
				logger.Info("message", "type", m)
				return d.DeletePeerFromCluster(context, m.Payload)
				break
			}
			return nil
		},
	)

	return consumer, err
}
