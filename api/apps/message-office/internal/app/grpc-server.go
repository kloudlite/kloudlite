package app

import (
	context "context"
	"fmt"
	"time"

	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/kloudlite/operator/pkg/kubectl"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"kloudlite.io/apps/message-office/internal/domain"
	"kloudlite.io/apps/message-office/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
)

type grpcServer struct {
	messages.UnimplementedMessageDispatchServiceServer
	logger logging.Logger

	producer  redpanda.Producer
	consumers map[string]redpanda.Consumer
	ev        *env.Env

	domain domain.Domain

	containerRegistryCli container_registry.ContainerRegistryClient
	k8sControllerCli     kubectl.ControllerClient

	resourceUpdatesCounter   int64
	infraUpdatesCounter      int64
	errorMessagesCounter     int64
	byocClientUpdatesCounter int64
}

// ReceiveErrors implements messages.MessageDispatchServiceServer
func (g *grpcServer) ReceiveErrors(server messages.MessageDispatchService_ReceiveErrorsServer) error {
	for {
		errorMsg, err := server.Recv()
		if err != nil {
			return err
		}

		g.errorMessagesCounter++
		g.logger.Infof("[%v] received error-on-apply message", g.errorMessagesCounter)
		g.logger.Infof("[%v] [error]: %s\n", g.errorMessagesCounter, errorMsg.Data)

		if err := g.domain.ValidateAccessToken(server.Context(), errorMsg.AccessToken, errorMsg.AccountName, errorMsg.ClusterName); err != nil {
			g.logger.Errorf(err, fmt.Sprintf("[%v] ERROR while validating access token", g.resourceUpdatesCounter))
			return err
		}

		po, err := g.producer.Produce(server.Context(), g.ev.KafkaTopicErrorOnApply, errorMsg.ClusterName, errorMsg.Data)
		if err != nil {
			g.logger.Errorf(err, fmt.Sprintf("[%v] ERROR while producing to topic (%s)", g.resourceUpdatesCounter, g.ev.KafkaTopicErrorOnApply))
			return err
		}
		g.logger.WithKV("topic", g.ev.KafkaTopicErrorOnApply).
			WithKV("parition", po.Partition).
			WithKV("offset", po.Offset).
			Infof("%v dispatched error-on-apply message", g.errorMessagesCounter)
	}
}

// GetAccessToken implements messages.MessageDispatchServiceServer
func (g *grpcServer) GetAccessToken(ctx context.Context, msg *messages.GetClusterTokenIn) (*messages.GetClusterTokenOut, error) {
	g.logger.Infof("request received for clustertoken: %s", msg.ClusterToken)
	defer func() {
		g.logger.Infof("request processed for clustertoken: %s", msg.ClusterToken)
	}()

	record, err := g.domain.GenAccessToken(ctx, msg.ClusterToken)
	if err != nil {
		return nil, err
	}

	var hu artifactsv1.HarborUserAccount
	if err := g.k8sControllerCli.Get(ctx, types.NamespacedName{Namespace: constants.NamespaceCore, Name: record.AccountName}, &hu); err != nil {
		return nil, err
	}

	var harborSecret corev1.Secret
	if err := g.k8sControllerCli.Get(ctx, types.NamespacedName{Namespace: constants.NamespaceCore, Name: hu.Spec.TargetSecret}, &harborSecret); err != nil {
		return nil, err
	}

	return &messages.GetClusterTokenOut{
		AccessToken:            record.AccessToken,
		HarborDockerConfigJson: string(harborSecret.Data[".dockerconfigjson"]),
	}, nil
}

func (g *grpcServer) createConsumer(ev *env.Env, topicName string) (redpanda.Consumer, error) {
	return redpanda.NewConsumer(ev.KafkaBrokers, ev.KafkaConsumerGroup, redpanda.ConsumerOpts{
		// SASLAuth: &redpanda.KafkaSASLAuth{
		// 	SASLMechanism: redpanda.ScramSHA256,
		// 	User:          ev.KafkaSaslUsername,
		// 	Password:      ev.KafkaSaslPassword,
		// },
		Logger: g.logger.WithName("g-consumer"),
	}, []string{topicName})
}

func (g grpcServer) SendActions(
	request *messages.StreamActionsRequest,
	server messages.MessageDispatchService_SendActionsServer,
) error {
	if err := g.domain.ValidateAccessToken(server.Context(), request.AccessToken, request.AccountName, request.ClusterName); err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s", request.AccountName, request.ClusterName)

	consumer, err := func() (redpanda.Consumer, error) {
		if c, ok := g.consumers[key]; ok {
			return c, nil
		}
		c, err := g.createConsumer(
			g.ev,
			common.GetKafkaTopicName(request.AccountName, request.ClusterName),
		)
		if err != nil {
			return nil, err
		}
		if err := c.Ping(server.Context()); err != nil {
			return nil, err
		}
		g.consumers[key] = c
		return c, nil
	}()
	if err != nil {
		return err
	}

	go func() {
		<-server.Context().Done()
		g.logger.Debugf("server context has been closed")
		delete(g.consumers, key)
		consumer.Close()
	}()

	defer func() {
		g.logger.Debugf("kafka consumer has been closed")
		delete(g.consumers, key)
		consumer.Close()
	}()

	consumer.StartConsumingSync(func(msg []byte, timeStamp time.Time, offset int64) error {
		g.logger.WithKV("timestamp", timeStamp).Infof("received message")
		defer func() {
			g.logger.WithKV("timestamp", timeStamp).Infof("processed message")
		}()
		return server.Send(&messages.Action{Data: msg})
	})
	return nil
}

func (g *grpcServer) ReceiveResourceUpdates(server messages.MessageDispatchService_ReceiveResourceUpdatesServer) error {
	for {
		statusMsg, err := server.Recv()
		if err != nil {
			return err
		}

		g.resourceUpdatesCounter++
		g.logger.Infof("[%v] received status update", g.resourceUpdatesCounter)

		if err = g.domain.ValidateAccessToken(server.Context(), statusMsg.AccessToken, statusMsg.AccountName, statusMsg.ClusterName); err != nil {
			g.logger.Errorf(err, fmt.Sprintf("[%v] ERROR while processing resource update", g.resourceUpdatesCounter))
			return err
		}

		po, err := g.producer.Produce(server.Context(), g.ev.KafkaTopicStatusUpdates, statusMsg.ClusterName, statusMsg.Message)
		if err != nil {
			g.logger.Errorf(err, fmt.Sprintf("[%v] ERROR while processing resource update", g.resourceUpdatesCounter))
			return err
		}
		g.logger.Infof("[%v] processed status update", g.resourceUpdatesCounter)
		g.logger.WithKV("topic", g.ev.KafkaTopicStatusUpdates).WithKV("parition", po.Partition).WithKV("offset", po.Offset).Infof("%v dispatched status updates", g.resourceUpdatesCounter)
	}
}

func (g *grpcServer) ReceiveBYOCClientUpdates(server messages.MessageDispatchService_ReceiveBYOCClientUpdatesServer) (err error) {
	for {
		clientUpdateMsg, err := server.Recv()
		if err != nil {
			return err
		}

		g.byocClientUpdatesCounter++
		g.logger.Infof("[%v] received byoc client update", g.byocClientUpdatesCounter)

		if err = g.domain.ValidateAccessToken(server.Context(), clientUpdateMsg.AccessToken, clientUpdateMsg.AccountName, clientUpdateMsg.ClusterName); err != nil {
			g.logger.Errorf(err, fmt.Sprintf("[%v] ERROR while processing BYOC Client update message", g.resourceUpdatesCounter))
			return err
		}

		po, err := g.producer.Produce(server.Context(), g.ev.KafkaTopicBYOCClientUpdates, clientUpdateMsg.ClusterName, clientUpdateMsg.Message)
		if err != nil {
			g.logger.Errorf(err, fmt.Sprintf("[%v] ERROR while processing BYOC Client update message", g.resourceUpdatesCounter))
			return err
		}
		g.logger.Infof("[%v] processed BYOC Client ClientUpdate", g.byocClientUpdatesCounter)
		g.logger.WithKV("topic", g.ev.KafkaTopicBYOCClientUpdates).
			WithKV("parition", po.Partition).
			WithKV("offset", po.Offset).
			Infof("%v dispatched byoc client updates", g.byocClientUpdatesCounter)
	}
}

// ReceiveInfraUpdates implements messages.MessageDispatchServiceServer
func (g *grpcServer) ReceiveInfraUpdates(
	server messages.MessageDispatchService_ReceiveInfraUpdatesServer,
) (err error) {
	for {
		statusMsg, err := server.Recv()
		if err != nil {
			return err
		}

		g.infraUpdatesCounter++
		g.logger.Infof("%v received infra update", g.infraUpdatesCounter)

		defer func() {
			if err != nil {
				g.logger.Errorf(
					err,
					fmt.Sprintf("[%v] ERROR while processing infra update", g.infraUpdatesCounter),
				)
				return
			}
			g.logger.Infof("[%v] processed infra update", g.infraUpdatesCounter)
		}()

		if err := g.domain.ValidateAccessToken(server.Context(), statusMsg.AccessToken, statusMsg.AccountName, statusMsg.ClusterName); err != nil {
			return err
		}

		po, err := g.producer.Produce(
			server.Context(),
			g.ev.KafkaTopicInfraUpdates,
			statusMsg.ClusterName,
			statusMsg.Message,
		)
		if err != nil {
			return err
		}
		g.logger.WithKV("topic", g.ev.KafkaTopicInfraUpdates).
			WithKV("parition", po.Partition).
			WithKV("offset", po.Offset).
			Infof("%v dispatched infra updates", g.infraUpdatesCounter)
	}
}
