package app

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/pkg/errors"

	"kloudlite.io/apps/message-office/internal/domain"
	"kloudlite.io/apps/message-office/internal/env"
	"kloudlite.io/common"
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

	resourceUpdatesCounter int64
	infraUpdatesCounter    int64
	errorMessagesCounter   int64
	clusterUpdatesCounter  int64
}

// func (g *grpcServer) GetDockerCredentials(ctx context.Context, in *messages.GetDockerCredentialsIn) (out *messages.GetDockerCredentialsOut, err error) {
// 	logger := g.logger.WithKV("accountName", in.AccountName, "clusterName", in.ClusterName)
// 	logger.Infof("request received for docker credentials")
// 	defer func() {
// 		if err != nil {
// 			logger.Errorf(err, "error occurred while processing for docker credentials")
// 			return
// 		}
// 		g.logger.Infof("request processed for docker credentials")
// 	}()
//
// 	if err := g.domain.ValidateAccessToken(ctx, in.AccessToken, in.AccountName, in.ClusterName); err != nil {
// 		return nil, err
// 	}
//
// 	ns, err := common.FindNamespaceForAccount(ctx, g.k8sControllerCli, in.AccountName)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var secretsList corev1.SecretList
// 	if err := g.k8sControllerCli.List(ctx, &secretsList, &client.ListOptions{
// 		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
// 			"kloudlite.io/secret.type": string(corev1.SecretTypeDockerConfigJson),
// 		}),
// 		Namespace: ns.Name,
// 		Limit:     0,
// 	}); err != nil {
// 		return nil, err
// 	}
//
// 	secrets := make(map[string]string, len(secretsList.Items))
// 	for i := range secretsList.Items {
// 		secrets[secretsList.Items[i].Name] = string(secretsList.Items[i].Data[".dockerconfigjson"])
// 	}
//
// 	return &messages.GetDockerCredentialsOut{DockerConfigJsonSecrets: secrets}, nil
// }

func (g *grpcServer) parseError(ctx context.Context, errMsg *messages.ErrorData) (err error) {
	g.errorMessagesCounter++
	logger := g.logger.WithKV("accountName", errMsg.AccountName).WithKV("cluster", errMsg.ClusterName)

	logger.Infof("[%v] received error-on-apply message", g.errorMessagesCounter)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[%v] (with ERROR) processed error-on-apply message", g.clusterUpdatesCounter))
			logger.Errorf(err)
			return
		}
		logger.Infof("[%v] processed error-on-apply message", g.infraUpdatesCounter)
	}()

	if err := g.domain.ValidateAccessToken(ctx, errMsg.AccessToken, errMsg.AccountName, errMsg.ClusterName); err != nil {
		return errors.Wrap(err, "while validating access token")
	}

	if _, err := g.producer.Produce(ctx, g.ev.KafkaTopicErrorOnApply, errMsg.ClusterName, errMsg.Data); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing to topic (%s)", g.ev.KafkaTopicErrorOnApply))
	}
	logger.Infof("[%v] dispatched error-on-apply message", g.errorMessagesCounter)
	return nil
}

// ReceiveErrors implements messages.MessageDispatchServiceServer
func (g *grpcServer) ReceiveErrors(server messages.MessageDispatchService_ReceiveErrorsServer) error {
	for {
		errorMsg, err := server.Recv()
		if err != nil {
			return err
		}
		_ = g.parseError(server.Context(), errorMsg)
	}
}

// GetAccessToken implements messages.MessageDispatchServiceServer
func (g *grpcServer) GetAccessToken(ctx context.Context, msg *messages.GetClusterTokenIn) (*messages.GetClusterTokenOut, error) {
	g.logger.Infof("request received for cluster-token (%q) exchange", msg.ClusterToken)

	record, err := g.domain.GenAccessToken(ctx, msg.ClusterToken)
	if err != nil {
		return nil, err
	}

	g.logger.Infof("SUCCESSFUL cluster-token (%q) exchange for account=%q, cluster=%q", msg.ClusterToken, record.AccountName, record.ClusterName)

	return &messages.GetClusterTokenOut{
		AccessToken: record.AccessToken,
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

func (g *grpcServer) SendActions(request *messages.StreamActionsRequest, server messages.MessageDispatchService_SendActionsServer) error {
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

	consumer.StartConsumingSync(func(msg []byte, timeStamp time.Time, offset int64) error {
		g.logger.WithKV("timestamp", timeStamp).Infof("received message")
		defer func() {
			g.logger.WithKV("timestamp", timeStamp).Infof("processed message")
		}()
		return server.Send(&messages.Action{Data: msg})
	})
	return nil
}

func (g *grpcServer) processResourceUpdate(ctx context.Context, msg *messages.ResourceUpdate) (err error) {
	g.resourceUpdatesCounter++

	logger := g.logger.WithKV("accountName", msg.AccountName).WithKV("clusterName", msg.ClusterName)
	logger.Infof("[%v] received resource status update", g.resourceUpdatesCounter)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[%v] (with ERROR) processed resource status update", g.clusterUpdatesCounter))
			logger.Errorf(err)
			return
		}
		logger.Infof("[%v] processed resource status update", g.infraUpdatesCounter)
	}()

	if err = g.domain.ValidateAccessToken(ctx, msg.AccessToken, msg.AccountName, msg.ClusterName); err != nil {
		return errors.Wrap(err, fmt.Sprintf("[%v] while validating accessToken", g.resourceUpdatesCounter))
	}

	if _, err := g.producer.Produce(ctx, g.ev.KafkaTopicStatusUpdates, msg.ClusterName, msg.Message); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing resource update to topic %q", g.ev.KafkaTopicStatusUpdates))
	}

	logger.Infof("[%v] dispatched status updates to topic %q", g.resourceUpdatesCounter, g.ev.KafkaTopicStatusUpdates)
	return nil
}

func (g *grpcServer) ReceiveResourceUpdates(server messages.MessageDispatchService_ReceiveResourceUpdatesServer) error {
	for {
		statusMsg, err := server.Recv()
		if err != nil {
			return err
		}
		_ = g.processResourceUpdate(server.Context(), statusMsg)
	}
}

func (g *grpcServer) processClusterUpdate(ctx context.Context, msg *messages.ClusterUpdate) (err error) {
	g.clusterUpdatesCounter++
	logger := g.logger.WithKV("accountName", msg.AccountName).WithKV("clusterName", msg.ClusterName)

	logger.Infof("[%v] received Cluster update", g.clusterUpdatesCounter)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[%v] (with ERROR) processed Cluster update", g.clusterUpdatesCounter))
			logger.Errorf(err)
			return
		}
		logger.Infof("[%v] processed Cluster update", g.infraUpdatesCounter)
	}()

	if err = g.domain.ValidateAccessToken(ctx, msg.AccessToken, msg.AccountName, msg.ClusterName); err != nil {
		return errors.Wrap(err, "while validating access token")
	}

	if _, err := g.producer.Produce(ctx, g.ev.KafkaTopicClusterUpdates, msg.ClusterName, msg.Message); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing message into kafka topic (%s) for ", g.ev.KafkaTopicClusterUpdates))
	}

	logger.Infof("%v dispatched cluster updates into topic=%q", g.clusterUpdatesCounter, g.ev.KafkaTopicClusterUpdates)
	return nil
}

func (g *grpcServer) ReceiveClusterUpdates(server messages.MessageDispatchService_ReceiveClusterUpdatesServer) (err error) {
	for {
		clientUpdateMsg, err := server.Recv()
		if err != nil {
			return err
		}

		_ = g.processClusterUpdate(server.Context(), clientUpdateMsg)
	}
}

func (g *grpcServer) processInfraUpdate(ctx context.Context, msg *messages.InfraUpdate) (err error) {
	g.infraUpdatesCounter++

	logger := g.logger.WithKV("accountName", msg.AccountName).WithKV("clusterName", msg.ClusterName)

	logger.Infof("[%v] received infra update", g.infraUpdatesCounter)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[%v] (with ERROR) processed infra update", g.infraUpdatesCounter))
			g.logger.Errorf(err)
			return
		}
		g.logger.Infof("[%v] processed infra update", g.infraUpdatesCounter)
	}()

	if err := g.domain.ValidateAccessToken(ctx, msg.AccessToken, msg.AccountName, msg.ClusterName); err != nil {
		return err
	}

	po, err := g.producer.Produce(ctx, g.ev.KafkaTopicInfraUpdates, msg.ClusterName, msg.Message)
	if err != nil {
		return err
	}

	g.logger.WithKV("topic", g.ev.KafkaTopicInfraUpdates).
		WithKV("partition", po.Partition).
		WithKV("offset", po.Offset).
		Infof("%v dispatched infra updates", g.infraUpdatesCounter)
	return nil
}

// ReceiveInfraUpdates implements messages.MessageDispatchServiceServer
func (g *grpcServer) ReceiveInfraUpdates(server messages.MessageDispatchService_ReceiveInfraUpdatesServer) (err error) {
	for {
		statusMsg, err := server.Recv()
		if err != nil {
			return err
		}
		_ = g.processInfraUpdate(server.Context(), statusMsg)
	}
}
