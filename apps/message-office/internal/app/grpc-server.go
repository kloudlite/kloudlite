package app

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	klErrors "github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"

	"github.com/kloudlite/api/pkg/messaging"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/api/pkg/nats"

	"github.com/kloudlite/api/apps/message-office/internal/domain"
	"github.com/kloudlite/api/apps/message-office/internal/env"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
)

type (
	UpdatesProducer messaging.Producer
	grpcServer      struct {
		messages.UnimplementedMessageDispatchServiceServer
		logger logging.Logger

		updatesProducer UpdatesProducer
		consumers       map[string]messaging.Consumer
		ev              *env.Env

		domain domain.Domain

		createConsumer func(ctx context.Context, accountName string, clusterName string) (messaging.Consumer, error)

		resourceUpdatesCounter int64
		infraUpdatesCounter    int64
		errorMessagesCounter   int64
		clusterUpdatesCounter  int64
	}
)

// Ping implements messages.MessageDispatchServiceServer.
func (*grpcServer) Ping(context.Context, *messages.Empty) (*messages.PingOutput, error) {
	return &messages.PingOutput{Ok: true}, nil
}

// mustEmbedUnimplementedMessageDispatchServiceServer implements messages.MessageDispatchServiceServer.
func (*grpcServer) mustEmbedUnimplementedMessageDispatchServiceServer() {
	panic("unimplemented")
}

func encodeAccessToken(accountName, clusterName, clusterToken string, tokenSecret string) string {
	info := fmt.Sprintf("account=%s;cluster=%s;cluster-token=%s", accountName, clusterName, clusterToken)

	fn.FxErrorHandler()

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

func (g *grpcServer) validateAndDecodeFromGrpcContext(grpcServerCtx context.Context, tokenSecret string) (accountName string, clusterName string, err error) {
	authToken := metadata.ValueFromIncomingContext(grpcServerCtx, "authorization")
	if len(authToken) != 1 {
		return "", "", errors.New("no authorization header passed")
	}

	if authToken[0] != g.ev.PlatformAccessToken {
		return validateAndDecodeAccessToken(authToken[0], tokenSecret)
	}

	splits := strings.Split(authToken[0], ";")
	for _, v := range splits {
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

func (g *grpcServer) ValidateAccessToken(ctx context.Context, msg *messages.ValidateAccessTokenIn) (*messages.ValidateAccessTokenOut, error) {
	logger := g.logger.WithKV("accountName", msg.AccountName).WithKV("cluster", msg.ClusterName)
	logger.Debugf("request received for access token validation")
	isValid := true
	defer func() {
		logger.Debugf("is access token valid? (%v)", isValid)
	}()

	if msg.AccessToken == g.ev.PlatformAccessToken {
		return &messages.ValidateAccessTokenOut{Valid: true}, nil
	}

	if _, _, err := validateAndDecodeAccessToken(msg.AccessToken, g.ev.TokenHashingSecret); err != nil {
		isValid = false
	}
	return &messages.ValidateAccessTokenOut{Valid: isValid}, nil
}

// FIXME: this should be split into 2 methods, one for console resource errors, and other for infra errors
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

	msgTopic := common.GetPlatformClusterMessagingTopic(accountName, clusterName, common.InfraReceiver, common.EventErrorOnApply)
	if err := g.updatesProducer.Produce(ctx, types.ProduceMsg{
		Subject: msgTopic,
		Payload: errMsg.Message,
	}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing to topic (%s)", msgTopic))
	}
	logger.Infof("[%v] dispatched error-on-apply message", g.errorMessagesCounter)
	return nil
}

// ReceiveErrors implements messages.MessageDispatchServiceServer
func (g *grpcServer) ReceiveErrors(server messages.MessageDispatchService_ReceiveErrorsServer) error {
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return klErrors.NewE(err)
	}
	for {
		errorMsg, err := server.Recv()
		if err != nil {
			return klErrors.NewE(err)
		}
		_ = g.parseError(server.Context(), accountName, clusterName, errorMsg)
	}
}

// GetAccessToken implements messages.MessageDispatchServiceServer
func (g *grpcServer) GetAccessToken(ctx context.Context, msg *messages.GetClusterTokenIn) (*messages.GetClusterTokenOut, error) {
	g.logger.Infof("request received for cluster-token (%q) exchange", msg.ClusterToken)

	ct, err := g.domain.GetClusterToken(ctx, msg.AccountName, msg.ClusterName)
	if err != nil {
		return nil, klErrors.NewE(err)
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

func (g *grpcServer) SendActions(request *messages.Empty, server messages.MessageDispatchService_SendActionsServer) error {
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return klErrors.NewE(err)
	}

	logger := g.logger.WithKV("accountName", accountName, "clusterName", clusterName)
	logger.Infof("request received for sending actions to cluster")
	defer func() {
		logger.Infof("stopped sending actions to cluster")
	}()

	key := fmt.Sprintf("%s/%s", accountName, clusterName)

	consumer, err := g.createConsumer(server.Context(), accountName, clusterName)
	if err != nil {
		return klErrors.NewE(err)
	}

	logger.Infof("consumer is available now")

	go func() {
		<-server.Context().Done()
		logger.Debugf("server context has been closed")
		delete(g.consumers, key)
		if err := consumer.Stop(context.TODO()); err != nil {
			logger.Errorf(err, "while stopping consumer")
		}
		logger.Infof("consumer is closed now")
	}()

	if err:=consumer.Consume(func(msg *types.ConsumeMsg) error {
		logger.WithKV("subject", msg.Subject).Infof("read message from consumer")
		defer func() {
			logger.WithKV("subject", msg.Subject).Infof("dispatched message to agent")
		}()
		return server.Send(&messages.Action{Message: msg.Payload})
	}, types.ConsumeOpts{
		OnError: func(error) error {
			logger.Infof("error occurrred on agent side, while parsing/applying the message, ignoring as we don't want to block the queue")
			return nil
		},
	}); err != nil {
		logger.Errorf(err, "while consuming messages from consumer")
	}

	return nil
}

func (g *grpcServer) processResourceUpdate(ctx context.Context, accountName string, clusterName string, msg *messages.ResourceUpdate) (err error) {
	g.resourceUpdatesCounter++

	logger := g.logger.WithKV("accountName", accountName).WithKV("clusterName", clusterName).WithKV("component", "console-resource-update")
	logger.Infof("[%v] received resource status update", g.resourceUpdatesCounter)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[%v] (with ERROR) processed resource status update", g.clusterUpdatesCounter))
			logger.Errorf(err)
			return
		}
		logger.Infof("[%v] processed resource status update", g.resourceUpdatesCounter)
	}()

	msgTopic := common.GetPlatformClusterMessagingTopic(accountName, clusterName, common.ConsoleReceiver, common.EventResourceUpdate)
	if err := g.updatesProducer.Produce(ctx, types.ProduceMsg{
		Subject: msgTopic,
		Payload: msg.Message,
	}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing resource update to topic %q", msgTopic))
	}

	logger.Infof("[%v] dispatched resource updates to topic %q", g.resourceUpdatesCounter, msgTopic)
	return nil
}

func (g *grpcServer) ReceiveResourceUpdates(server messages.MessageDispatchService_ReceiveResourceUpdatesServer) error {
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return klErrors.NewE(err)
	}
	for {
		statusMsg, err := server.Recv()
		if err != nil {
			return klErrors.NewE(err)
		}
		if err := g.processResourceUpdate(server.Context(), accountName, clusterName, statusMsg); err != nil {
			return err
		}
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

	msgTopic := common.GetPlatformClusterMessagingTopic(accountName, clusterName, common.InfraReceiver, common.EventResourceUpdate)
	if err := g.updatesProducer.Produce(ctx, types.ProduceMsg{
		Subject: msgTopic,
		Payload: msg.Message,
	}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing resource update to topic %q", msgTopic))
	}

	logger.Infof("%v dispatched cluster updates into topic=%q", g.clusterUpdatesCounter, msgTopic)
	return nil
}

func (g *grpcServer) ReceiveClusterUpdates(server messages.MessageDispatchService_ReceiveClusterUpdatesServer) (err error) {
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return klErrors.NewE(err)
	}
	for {
		clientUpdateMsg, err := server.Recv()
		if err != nil {
			return klErrors.NewE(err)
		}

		_ = g.processClusterUpdate(server.Context(), accountName, clusterName, clientUpdateMsg)
	}
}

func (g *grpcServer) processInfraUpdate(ctx context.Context, accountName string, clusterName string, msg *messages.InfraUpdate) (err error) {
	g.infraUpdatesCounter++
	logger := g.logger.WithKV("accountName", accountName).WithKV("clusterName", clusterName).WithKV("component", "infra-resource-update")

	logger.Infof("[%v] received infra update", g.infraUpdatesCounter)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[%v] (with ERROR) processed infra update", g.infraUpdatesCounter))
			logger.Errorf(err)
			return
		}
		logger.Infof("[%v] processed infra update", g.infraUpdatesCounter)
	}()

	msgTopic := common.GetPlatformClusterMessagingTopic(accountName, clusterName, common.InfraReceiver, common.EventResourceUpdate)
	if err := g.updatesProducer.Produce(ctx, types.ProduceMsg{
		Subject: msgTopic,
		Payload: msg.Message,
	}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing resource update to topic %q", msgTopic))
	}

	logger.Infof("[%v] processed infra update", g.infraUpdatesCounter)
	return nil
}

func (g *grpcServer) processCRUpdate(ctx context.Context, accountName string, clusterName string, msg *messages.ContainerRegistryUpdate) (err error) {
	g.infraUpdatesCounter++
	logger := g.logger.WithKV("accountName", accountName).WithKV("clusterName", clusterName).WithKV("component", "container-registry-update")

	logger.Infof("[%v] received cr update", g.infraUpdatesCounter)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[%v] (with ERROR) processed cr update", g.infraUpdatesCounter))
			logger.Errorf(err)
			return
		}
		logger.Infof("[%v] processed cr update", g.infraUpdatesCounter)
	}()

	msgTopic := common.GetPlatformClusterMessagingTopic(accountName, clusterName, common.ContainerRegistryReceiver, common.EventResourceUpdate)
	if err := g.updatesProducer.Produce(ctx, types.ProduceMsg{
		Subject: msgTopic,
		Payload: msg.Message,
	}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing resource update to topic %q", msgTopic))
	}

	logger.Infof("[%v] processed cr update", g.infraUpdatesCounter)
	return nil
}


// ReceiveInfraUpdates implements messages.MessageDispatchServiceServer
func (g *grpcServer) ReceiveInfraUpdates(server messages.MessageDispatchService_ReceiveInfraUpdatesServer) (err error) {
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return klErrors.NewE(err)
	}
	for {
		statusMsg, err := server.Recv()
		if err != nil {
			return klErrors.NewE(err)
		}
		_ = g.processInfraUpdate(server.Context(), accountName, clusterName, statusMsg)
	}
}

func (g *grpcServer) ReceiveContainerRegistryUpdates(server messages.MessageDispatchService_ReceiveContainerRegistryUpdatesServer) error{
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return klErrors.NewE(err)
	}
	for {
		statusMsg, err := server.Recv()
		if err != nil {
			return klErrors.NewE(err)
		}
		_ = g.processCRUpdate(server.Context(), accountName, clusterName, statusMsg)
	}
}



func NewMessageOfficeServer(producer UpdatesProducer, jc *nats.JetstreamClient, ev *env.Env, d domain.Domain, logger logging.Logger) (messages.MessageDispatchServiceServer, error) {
	return &grpcServer{
		UnimplementedMessageDispatchServiceServer: messages.UnimplementedMessageDispatchServiceServer{},
		logger:          logger,
		updatesProducer: producer,
		consumers:       map[string]messaging.Consumer{},
		ev:              ev,
		domain:          d,
		createConsumer: func(ctx context.Context, accountName string, clusterName string) (messaging.Consumer, error) {
			name := fmt.Sprintf("tenant-consumer-for-account-%s-cluster-%s", accountName, clusterName)

			return msg_nats.NewJetstreamConsumer(ctx, jc, msg_nats.JetstreamConsumerArgs{
				Stream: ev.NatsStream,
				ConsumerConfig: msg_nats.ConsumerConfig{
					Name:        name,
					Durable:     name,
					Description: "this consumer consumes messages from platform, and dispatches them to the tenant cluster via kloudlite agent",
					FilterSubjects: []string{
						common.GetTenantClusterMessagingTopic(accountName, clusterName),
					},
				},
			})
		},
	}, nil
}
