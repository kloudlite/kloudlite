package app

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"

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

func encodeAccessToken(accountName, clusterName, clusterToken string, tokenSecret string) string {
	info := fmt.Sprintf("account=%s;cluster=%s;cluster-token=%s", accountName, clusterName, clusterToken)

	h := sha256.New()
	h.Write([]byte(info + tokenSecret))
	sum := fmt.Sprintf("%x", h.Sum(nil))

	info += fmt.Sprintf(";sha256sum=%s", sum)

	return base64.StdEncoding.EncodeToString([]byte(info))
}

func validateAndDecodeAccessToken(accessToken string, tokenSecret string) (accountName string, clusterName string, err error) {
	b, err := base64.StdEncoding.DecodeString(accessToken)
	if err != nil {
		return "", "", errors.Wrap(err, "invalid access token, incorrect format")
	}

	info := string(b)

	sp := strings.SplitN(info, ";sha256sum=", 2)
	if len(sp) != 2 {
		return "", "", errors.New("invalid access token, incorrect format")
	}
	data := sp[0]
	sum := sp[1]

	h := sha256.New()
	h.Write([]byte(data + tokenSecret))
	calculatedSum := fmt.Sprintf("%x", h.Sum(nil))

	if sum != calculatedSum {
		return "", "", errors.New("invalid access token, checksum mismatch")
	}

	s := strings.SplitN(data, ";", 3)
	if len(s) != 3 {
		return "", "", errors.New("invalid access token, incorrect data format")
	}

	for _, v := range s {
		sp := strings.SplitN(v, "=", 2)
		if len(sp) != 2 {
			return "", "", errors.New("invalid access token, incorrect data format")
		}
		if sp[0] == "account" {
			accountName = sp[1]
		}
		if sp[0] == "cluster" {
			clusterName = sp[1]
		}
	}

	return accountName, clusterName, nil
}

func validateAndDecodeFromGrpcContext(ctx context.Context, tokenSecret string) (accountName string, clusterName string, err error) {
	authToken := metadata.ValueFromIncomingContext(ctx, "authorization")
	if len(authToken) != 1 {
		return "", "", errors.New("no authorization header passed")
	}
	return validateAndDecodeAccessToken(authToken[0], tokenSecret)
}

func (g *grpcServer) ValidateAccessToken(ctx context.Context, msg *messages.ValidateAccessTokenIn) (*messages.ValidateAccessTokenOut, error) {
	logger := g.logger.WithKV("accountName", msg.AccountName).WithKV("cluster", msg.ClusterName)
	logger.Infof("request received for access token validation")
	isValid := true
	defer func() {
		logger.Infof("is access token valid? (%v)", isValid)
	}()

	if _, _, err := validateAndDecodeAccessToken(msg.AccessToken, g.ev.TokenHashingSecret); err != nil {
		isValid = false
	}
	return &messages.ValidateAccessTokenOut{Valid: isValid}, nil
}

func (g *grpcServer) parseError(ctx context.Context, accountName string, clusterName string, errMsg *messages.ErrorData) (err error) {
	g.errorMessagesCounter++
	logger := g.logger.WithKV("accountName", accountName).WithKV("cluster", clusterName)

	logger.Infof("[%v] received error-on-apply message", g.errorMessagesCounter)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[%v] (with ERROR) processed error-on-apply message", g.clusterUpdatesCounter))
			logger.Errorf(err)
			return
		}
		logger.Infof("[%v] processed error-on-apply message", g.infraUpdatesCounter)
	}()

	if _, err := g.producer.Produce(ctx, g.ev.KafkaTopicErrorOnApply, clusterName, errMsg.Message); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing to topic (%s)", g.ev.KafkaTopicErrorOnApply))
	}
	logger.Infof("[%v] dispatched error-on-apply message", g.errorMessagesCounter)
	return nil
}

// ReceiveErrors implements messages.MessageDispatchServiceServer
func (g *grpcServer) ReceiveErrors(server messages.MessageDispatchService_ReceiveErrorsServer) error {
	accountName, clusterName, err := validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return err
	}
	for {
		errorMsg, err := server.Recv()
		if err != nil {
			return err
		}
		_ = g.parseError(server.Context(), accountName, clusterName, errorMsg)
	}
}

// GetAccessToken implements messages.MessageDispatchServiceServer
func (g *grpcServer) GetAccessToken(ctx context.Context, msg *messages.GetClusterTokenIn) (*messages.GetClusterTokenOut, error) {
	g.logger.Infof("request received for cluster-token (%q) exchange", msg.ClusterToken)

	ct, err := g.domain.GetClusterToken(ctx, msg.AccountName, msg.ClusterName)
	if err != nil {
		return nil, err
	}
	if ct != msg.ClusterToken {
		return nil, errors.New("invalid cluster-token,account-name,cluster-name triplet")
	}

	s := encodeAccessToken(msg.AccountName, msg.ClusterName, msg.ClusterToken, g.ev.TokenHashingSecret)
	g.logger.Infof("SUCCESSFUL cluster-token exchange for account=%q, cluster=%q", msg.ClusterToken, msg.AccountName, msg.ClusterName)

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

func (g *grpcServer) SendActions(request *messages.Empty, server messages.MessageDispatchService_SendActionsServer) error {
	accountName, clusterName, err := validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return err
	}

	logger := g.logger.WithKV("accountName", accountName, "clusterName", clusterName)
	logger.Infof("request received for sending actions to cluster")
	defer func() {
		logger.Infof("stopping sending actions to cluster")
	}()

	key := fmt.Sprintf("%s/%s", accountName, clusterName)

	consumer, err := func() (redpanda.Consumer, error) {
		if c, ok := g.consumers[key]; ok {
			return c, nil
		}
		c, err := g.createConsumer(g.ev, common.GetKafkaTopicName(accountName, clusterName))
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
		return server.Send(&messages.Action{Message: msg})
	})
	return nil
}

func (g *grpcServer) processResourceUpdate(ctx context.Context, accountName string, clusterName string, msg *messages.ResourceUpdate) (err error) {
	g.resourceUpdatesCounter++

	logger := g.logger.WithKV("accountName", accountName).WithKV("clusterName", clusterName)
	logger.Infof("[%v] received resource status update", g.resourceUpdatesCounter)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[%v] (with ERROR) processed resource status update", g.clusterUpdatesCounter))
			logger.Errorf(err)
			return
		}
		logger.Infof("[%v] processed resource status update", g.infraUpdatesCounter)
	}()

	if _, err := g.producer.Produce(ctx, g.ev.KafkaTopicStatusUpdates, clusterName, msg.Message); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing resource update to topic %q", g.ev.KafkaTopicStatusUpdates))
	}

	logger.Infof("[%v] dispatched status updates to topic %q", g.resourceUpdatesCounter, g.ev.KafkaTopicStatusUpdates)
	return nil
}

func (g *grpcServer) ReceiveResourceUpdates(server messages.MessageDispatchService_ReceiveResourceUpdatesServer) error {
	accountName, clusterName, err := validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return err
	}
	for {
		statusMsg, err := server.Recv()
		if err != nil {
			return err
		}
		_ = g.processResourceUpdate(server.Context(), accountName, clusterName, statusMsg)
	}
}

func (g *grpcServer) processClusterUpdate(ctx context.Context, accountName string, clusterName string, msg *messages.ClusterUpdate) (err error) {
	g.clusterUpdatesCounter++
	logger := g.logger.WithKV("accountName", accountName).WithKV("clusterName", clusterName)

	logger.Infof("[%v] received Cluster update", g.clusterUpdatesCounter)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[%v] (with ERROR) processed Cluster update", g.clusterUpdatesCounter))
			logger.Errorf(err)
			return
		}
		logger.Infof("[%v] processed Cluster update", g.infraUpdatesCounter)
	}()

	if _, err := g.producer.Produce(ctx, g.ev.KafkaTopicClusterUpdates, clusterName, msg.Message); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing message into kafka topic (%s) for ", g.ev.KafkaTopicClusterUpdates))
	}

	logger.Infof("%v dispatched cluster updates into topic=%q", g.clusterUpdatesCounter, g.ev.KafkaTopicClusterUpdates)
	return nil
}

func (g *grpcServer) ReceiveClusterUpdates(server messages.MessageDispatchService_ReceiveClusterUpdatesServer) (err error) {
	accountName, clusterName, err := validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return err
	}
	for {
		clientUpdateMsg, err := server.Recv()
		if err != nil {
			return err
		}

		_ = g.processClusterUpdate(server.Context(), accountName, clusterName, clientUpdateMsg)
	}
}

func (g *grpcServer) processInfraUpdate(ctx context.Context, accountName string, clusterName string, msg *messages.InfraUpdate) (err error) {
	g.infraUpdatesCounter++
	logger := g.logger.WithKV("accountName", accountName).WithKV("clusterName", clusterName)

	logger.Infof("[%v] received infra update", g.infraUpdatesCounter)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[%v] (with ERROR) processed infra update", g.infraUpdatesCounter))
			g.logger.Errorf(err)
			return
		}
		g.logger.Infof("[%v] processed infra update", g.infraUpdatesCounter)
	}()

	po, err := g.producer.Produce(ctx, g.ev.KafkaTopicInfraUpdates, clusterName, msg.Message)
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
	accountName, clusterName, err := validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return err
	}
	for {
		statusMsg, err := server.Recv()
		if err != nil {
			return err
		}
		_ = g.processInfraUpdate(server.Context(), accountName, clusterName, statusMsg)
	}
}
