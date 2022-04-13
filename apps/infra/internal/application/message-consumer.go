package application

import (
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
		func(topic string, action messaging.Message) error {
			var _d struct {
				Type string
			}
			action.Unmarshal(&_d)
			switch _d.Type {
			case "setup-cluster":
				var m struct {
					Type    string
					Payload domain.SetupClusterAction
				}
				action.Unmarshal(&m)
				return d.CreateCluster(m.Payload)
				break
			case "update-cluster":
				var m struct {
					Type    string
					Payload domain.UpdateClusterAction
				}
				action.Unmarshal(&m)
				return d.UpdateCluster(m.Payload)
				break
			case "delete-cluster":
				var m struct {
					Type    string
					Payload domain.DeleteClusterAction
				}
				action.Unmarshal(&m)
				return d.DeleteCluster(m.Payload)
				break
			case "add-peer":
				var m struct {
					Type    string
					Payload domain.AddPeerAction
				}
				action.Unmarshal(&m)
				return d.AddPeerToCluster(m.Payload)
				break
			case "delete-peer":
				var m struct {
					Type    string
					Payload domain.DeletePeerAction
				}
				action.Unmarshal(&m)
				return d.DeletePeerFromCluster(m.Payload)
				break
			}
			return nil
		},
	)

	return consumer, err
}
