package application

import (
	"context"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

func fxConsumer(env *InfraEnv, mc messaging.KafkaClient, d domain.Domain, logger logger.Logger) (messaging.Consumer, error) {
	consumer, err := messaging.NewKafkaConsumer(
		mc,
		[]string{env.KafkaInfraTopic},
		env.KafkaGroupId,
		logger,
		func(context context.Context, topic string, action messaging.Message) error {
			var _d struct {
				Type string
				//Payload map[string]interface{}
			}
			action.Unmarshal(&_d)
			//if true {
			//	logger.Info("message", "type", _d)
			//	return nil
			//}
			switch _d.Type {
			case "setup-cluster":
				var m struct {
					Type    string
					Payload domain.SetupClusterAction
				}
				action.Unmarshal(&m)
				logger.Info("message", "type", m)
				return d.CreateCluster(m.Payload)
				break
			case "update-cluster":
				var m struct {
					Type    string
					Payload domain.UpdateClusterAction
				}
				action.Unmarshal(&m)
				logger.Info("message", "type", m)
				return d.UpdateCluster(m.Payload)
				break
			case "delete-cluster":
				var m struct {
					Type    string
					Payload domain.DeleteClusterAction
				}
				action.Unmarshal(&m)
				logger.Info("message", "type", m)
				return d.DeleteCluster(m.Payload)
				break
			case "add-peer":
				var m struct {
					Type    string
					Payload domain.AddPeerAction
				}
				action.Unmarshal(&m)
				logger.Info("message", "type", m)
				return d.AddPeerToCluster(m.Payload)
				break
			case "delete-peer":
				var m struct {
					Type    string
					Payload domain.DeletePeerAction
				}
				action.Unmarshal(&m)
				logger.Info("message", "type", m)
				return d.DeletePeerFromCluster(m.Payload)
				break
			}
			return nil
		},
	)

	return consumer, err
}
