package app

import (
	"fmt"
	"time"

	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
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
}

func (g *grpcServer) createConsumer(ev *env.Env, topicName string) (redpanda.Consumer, error) {
	return redpanda.NewConsumer(ev.KafkaBrokers, ev.KafkaConsumerGroup, redpanda.ConsumerOpts{
		SASLAuth: &redpanda.KafkaSASLAuth{
			SASLMechanism: redpanda.ScramSHA256,
			User:          ev.KafkaSaslUsername,
			Password:      ev.KafkaSaslPassword,
		},
		Logger: g.logger.WithName("g-consumer"),
	}, []string{topicName})
}

func (g grpcServer) SendActions(request *messages.StreamActionsRequest, server messages.MessageDispatchService_SendActionsServer) error {
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

	defer func() {
		delete(g.consumers, key)
		consumer.Close()
	}()

	consumer.StartConsuming(func(msg []byte, timeStamp time.Time, offset int64) error {
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
		if _, err := g.producer.Produce(server.Context(), "kl-status-updates", statusMsg.ClusterName, statusMsg.StatusUpdateMessage); err != nil {
			return err
		}
	}
}
