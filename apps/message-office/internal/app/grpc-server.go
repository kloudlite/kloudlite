package app

import (
	context "context"
	"fmt"
	"time"

	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"kloudlite.io/apps/message-office/internal/domain"
	"kloudlite.io/apps/message-office/internal/env"
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

	statusUpdatesCounter int64
	infraUpdatesCounter  int64
	errorMessagesCounter int64
}

// ReceiveErrors implements messages.MessageDispatchServiceServer
func (g *grpcServer) ReceiveErrors(server messages.MessageDispatchService_ReceiveErrorsServer) error {
	for {
		errorMsg, err := server.Recv()
		if err != nil {
			return err
		}

		g.errorMessagesCounter++
		g.logger.Infof("%v received error-on-apply message", g.errorMessagesCounter)

		if err := g.domain.ValidationAccessToken(server.Context(), errorMsg.AccessToken, errorMsg.AccountName, errorMsg.ClusterName); err != nil {
			return err
		}

		po, err := g.producer.Produce(server.Context(), g.ev.KafkaTopicErrorOnApply, errorMsg.ClusterName, errorMsg.Data)
		if err != nil {
			return err
		}
		g.logger.WithKV("topic", g.ev.KafkaTopicInfraUpdates).WithKV("parition", po.Partition).WithKV("offset", po.Offset).Infof("%v dispatched error-on-apply messages", g.infraUpdatesCounter)
	}
}

// GetAccessToken implements messages.MessageDispatchServiceServer
func (g *grpcServer) GetAccessToken(ctx context.Context, msg *messages.GetClusterTokenIn) (*messages.GetClusterTokenOut, error) {
	g.logger.Infof("request received for clustertoken: %s", msg.ClusterToken)
	defer func() {
		g.logger.Infof("request processed for clustertoken: %s", msg.ClusterToken)
	}()
	s, err := g.domain.GenAccessToken(ctx, msg.ClusterToken)
	if err != nil {
		return nil, err
	}
	return &messages.GetClusterTokenOut{
		AccessToken: s,
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

func (g grpcServer) SendActions(request *messages.StreamActionsRequest, server messages.MessageDispatchService_SendActionsServer) error {
	if err := g.domain.ValidationAccessToken(server.Context(), request.AccessToken, request.AccountName, request.ClusterName); err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s", request.AccountName, request.ClusterName)

	consumer, err := func() (redpanda.Consumer, error) {
		if c, ok := g.consumers[key]; ok {
			return c, nil
		}
		c, err := g.createConsumer(g.ev, fmt.Sprintf("clus-%s-%s-incoming", request.AccountName, request.ClusterName))
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

func (g grpcServer) ReceiveStatusMessages(server messages.MessageDispatchService_ReceiveStatusMessagesServer) error {
	for {
		statusMsg, err := server.Recv()
		if err != nil {
			return err
		}

		g.statusUpdatesCounter++
		g.logger.Infof("%v received status update", g.statusUpdatesCounter)

		if err := g.domain.ValidationAccessToken(server.Context(), statusMsg.AccessToken, statusMsg.AccountName, statusMsg.ClusterName); err != nil {
			return err
		}

		po, err := g.producer.Produce(server.Context(), g.ev.KafkaTopicStatusUpdates, statusMsg.ClusterName, statusMsg.StatusUpdateMessage)
		if err != nil {
			return err
		}
		g.logger.WithKV("topic", g.ev.KafkaTopicStatusUpdates).WithKV("parition", po.Partition).WithKV("offset", po.Offset).Infof("%v dispatched status updates", g.statusUpdatesCounter)
	}
}

// ReceiveInfraUpdates implements messages.MessageDispatchServiceServer
func (g *grpcServer) ReceiveInfraUpdates(server messages.MessageDispatchService_ReceiveInfraUpdatesServer) error {
	for {
		statusMsg, err := server.Recv()
		if err != nil {
			return err
		}

		g.infraUpdatesCounter++
		g.logger.Infof("%v received infra update", g.statusUpdatesCounter)
		if err := g.domain.ValidationAccessToken(server.Context(), statusMsg.AccessToken, statusMsg.AccountName, statusMsg.ClusterName); err != nil {
			return err
		}

		po, err := g.producer.Produce(server.Context(), g.ev.KafkaTopicInfraUpdates, statusMsg.ClusterName, statusMsg.InfraUpdateMessage)
		if err != nil {
			return err
		}
		g.logger.WithKV("topic", g.ev.KafkaTopicInfraUpdates).WithKV("parition", po.Partition).WithKV("offset", po.Offset).Infof("%v dispatched infra updates", g.infraUpdatesCounter)
	}
}
