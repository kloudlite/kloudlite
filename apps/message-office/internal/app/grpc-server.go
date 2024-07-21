package app

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kloudlite/api/grpc-interfaces/infra"
	klErrors "github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/grpc"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	msgOfficeT "github.com/kloudlite/api/apps/message-office/types"
	"github.com/kloudlite/api/pkg/messaging"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/api/pkg/nats"

	"github.com/kloudlite/api/apps/message-office/internal/domain"
	"github.com/kloudlite/api/apps/message-office/internal/env"
	fn "github.com/kloudlite/api/pkg/functions"
)

type (
	UpdatesProducer messaging.Producer
	InfraGRPCClient grpc.Client

	grpcServer struct {
		messages.UnimplementedMessageDispatchServiceServer
		logger *slog.Logger

		infraClient infra.InfraClient

		updatesProducer UpdatesProducer
		consumers       map[string]messaging.Consumer
		ev              *env.Env

		domain domain.Domain

		createConsumer func(ctx context.Context, accountName string, clusterName string) (messaging.Consumer, error)
	}
)

// ReceiveConsoleResourceUpdate implements messages.MessageDispatchServiceServer.
func (g *grpcServer) ReceiveConsoleResourceUpdate(ctx context.Context, msg *messages.ResourceUpdate) (_ *messages.Empty, err error) {
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(ctx, g.ev.TokenHashingSecret)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	logger := g.logger.With("accountName", accountName, "cluster", "GVK", msg.Gvk, "NN", fmt.Sprintf("%s/%s", msg.Namespace, msg.Name), clusterName, "for", common.ConsoleReceiver, "request-id", fn.UUID())
	logger.Debug("RECEIVED resource update")
	defer func() {
		if err != nil {
			logger.Error("FAILED resource update", "took", fmt.Sprintf("%.2fs", time.Since(start).Seconds()), "err", err)
			return
		}

		logger.Info("DISPATCHED resource update", "took", fmt.Sprintf("%.2fs", time.Since(start).Seconds()))
	}()

	if err := dispatchResourceUpdate(ctx, common.ConsoleReceiver, ResourceUpdateArgs{
		logger:          logger,
		updatesProducer: g.updatesProducer,

		AccountName: accountName,
		ClusterName: clusterName,
		Message:     msg,
	}); err != nil {
		return nil, err
	}

	return &messages.Empty{}, nil
}

// ReceiveContainerRegistryUpdate implements messages.MessageDispatchServiceServer.
func (g *grpcServer) ReceiveContainerRegistryUpdate(ctx context.Context, msg *messages.ResourceUpdate) (_ *messages.Empty, err error) {
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(ctx, g.ev.TokenHashingSecret)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	logger := g.logger.With("accountName", accountName, "cluster", "GVK", msg.Gvk, "NN", fmt.Sprintf("%s/%s", msg.Namespace, msg.Name), clusterName, "for", common.ContainerRegistryReceiver, "request-id", fn.UUID())
	logger.Debug("RECEIVED resource update")
	defer func() {
		if err != nil {
			logger.Error("FAILED resource update", "took", fmt.Sprintf("%.2fs", time.Since(start).Seconds()), "err", err)
			return
		}
		logger.Info("DISPATCHED resource update", "took", fmt.Sprintf("%.2fs", time.Since(start).Seconds()))
	}()

	if err := dispatchResourceUpdate(ctx, common.ContainerRegistryReceiver, ResourceUpdateArgs{
		logger:          logger,
		updatesProducer: g.updatesProducer,

		AccountName: accountName,
		ClusterName: clusterName,
		Message:     msg,
	}); err != nil {
		return nil, err
	}

	return &messages.Empty{}, nil
}

// ReceiveError implements messages.MessageDispatchServiceServer.
func (g *grpcServer) ReceiveError(ctx context.Context, msg *messages.ErrorData) (*messages.Empty, error) {
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(ctx, g.ev.TokenHashingSecret)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	logger := g.logger.With("accountName", accountName, "cluster", "GVK", msg.Gvk, "NN", fmt.Sprintf("%s/%s", msg.Namespace, msg.Name), clusterName, "for", common.ContainerRegistryReceiver, "request-id", fn.UUID())
	logger.Debug("RECEIVED error-on-apply update")
	defer func() {
		if err != nil {
			logger.Error("FAILED error-on-apply update", "took", fmt.Sprintf("%.2fs", time.Since(start).Seconds()), "err", err)
			return
		}
		logger.Info("DISPATCHED error-on-apply update", "took", fmt.Sprintf("%.2fs", time.Since(start).Seconds()))
	}()

	if err := processError(ctx, ProcessErrorArgs{
		logger:          logger,
		updatesProducer: g.updatesProducer,

		AccountName: accountName,
		ClusterName: clusterName,
		Error:       msg,
	}); err != nil {
		return nil, err
	}

	return &messages.Empty{}, nil
}

// ReceiveInfraResourceUpdate implements messages.MessageDispatchServiceServer.
func (g *grpcServer) ReceiveInfraResourceUpdate(ctx context.Context, msg *messages.ResourceUpdate) (_ *messages.Empty, err error) {
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(ctx, g.ev.TokenHashingSecret)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	logger := g.logger.With("accountName", accountName, "cluster", "GVK", msg.Gvk, "NN", fmt.Sprintf("%s/%s", msg.Namespace, msg.Name), clusterName, "for", common.ContainerRegistryReceiver, "request-id", fn.UUID())
	logger.Debug("RECEIVED resource update")
	defer func() {
		if err != nil {
			logger.Error("FAILED resource update", "took", fmt.Sprintf("%.2fs", time.Since(start).Seconds()), "err", err)
			return
		}
		logger.Info("DISPATCHED resource update", "took", fmt.Sprintf("%.2fs", time.Since(start).Seconds()))
	}()

	if err := dispatchResourceUpdate(ctx, common.InfraReceiver, ResourceUpdateArgs{
		logger:          logger,
		updatesProducer: g.updatesProducer,

		AccountName: accountName,
		ClusterName: clusterName,
		Message:     msg,
	}); err != nil {
		return nil, err
	}

	return &messages.Empty{}, nil
}

// Ping implements messages.MessageDispatchServiceServer.
func (*grpcServer) Ping(context.Context, *messages.Empty) (*messages.PingOutput, error) {
	return &messages.PingOutput{Ok: true}, nil
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
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(ctx, g.ev.TokenHashingSecret)
	if err != nil {
		return nil, err
	}

	g.logger.With("account", accountName).With("cluster", clusterName).Info("validated access token")
	return &messages.ValidateAccessTokenOut{Valid: true}, nil
}

type ProcessErrorArgs struct {
	logger          *slog.Logger
	updatesProducer UpdatesProducer

	AccountName string
	ClusterName string
	Error       *messages.ErrorData
}

func processError(ctx context.Context, args ProcessErrorArgs) (err error) {
	b, err := msgOfficeT.MarshalErrMessage(msgOfficeT.ErrMessage{
		AccountName: args.AccountName,
		ClusterName: args.ClusterName,
		Error:       args.Error.Message,
	})
	if err != nil {
		return errors.Wrap(err, "marshaling resource update")
	}

	subjectParams := common.ReceiveFromAgentArgs{
		AccountName: args.AccountName,
		ClusterName: args.ClusterName,
		GVK:         args.Error.Gvk,
		Namespace:   args.Error.Namespace,
		Name:        args.Error.Name,
	}

	msgTopic := common.ReceiveFromAgentSubjectName(subjectParams, common.InfraReceiver, common.EventErrorOnApply)
	if err := args.updatesProducer.Produce(ctx, types.ProduceMsg{Subject: msgTopic, Payload: b}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing to topic (%s)", msgTopic))
	}

	args.logger.Debug("dispatched error-on-apply message", "subject", msgTopic, "to", common.InfraReceiver)

	msgTopic = common.ReceiveFromAgentSubjectName(subjectParams, common.ConsoleReceiver, common.EventErrorOnApply)
	if err := args.updatesProducer.Produce(ctx, types.ProduceMsg{Subject: msgTopic, Payload: b}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing to topic (%s)", msgTopic))
	}
	args.logger.Debug("dispatched error-on-apply message", "subject", msgTopic, "to", common.ConsoleReceiver)

	msgTopic = common.ReceiveFromAgentSubjectName(subjectParams, common.ContainerRegistryReceiver, common.EventErrorOnApply)
	if err := args.updatesProducer.Produce(ctx, types.ProduceMsg{Subject: msgTopic, Payload: b}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("while producing to topic (%s)", msgTopic))
	}
	args.logger.Debug("dispatched error-on-apply message", "subject", msgTopic, "to", common.ContainerRegistryReceiver)

	return nil
}

// GetAccessToken implements messages.MessageDispatchServiceServer
func (g *grpcServer) GetAccessToken(ctx context.Context, msg *messages.GetAccessTokenIn) (*messages.GetAccessTokenOut, error) {
	g.logger.Info("request received for cluster-token (%q) exchange", msg.ClusterToken)

	ct, err := g.domain.FindClusterToken(ctx, msg.ClusterToken)
	if err != nil {
		return nil, klErrors.NewE(err)
	}
	if ct == nil {
		return nil, errors.New("invalid cluster token")
	}

	s := encodeAccessToken(ct.AccountName, ct.ClusterName, msg.ClusterToken, g.ev.TokenHashingSecret)
	g.logger.Info("SUCCESSFUL cluster-token exchange for account=%q, cluster=%q", ct.AccountName, ct.ClusterName)

	return &messages.GetAccessTokenOut{
		ProtocolVersion: g.ev.GrpcMessageProtocolVersion,
		AccountName:     ct.AccountName,
		ClusterName:     ct.ClusterName,
		AccessToken:     s,
	}, nil
}

func (g *grpcServer) SendActions(request *messages.Empty, server messages.MessageDispatchService_SendActionsServer) error {
	accountName, clusterName, err := g.validateAndDecodeFromGrpcContext(server.Context(), g.ev.TokenHashingSecret)
	if err != nil {
		return klErrors.NewE(err)
	}

	logger := g.logger.With("accountName", accountName, "clusterName", clusterName)
	logger.Info("request received for sending actions to cluster")
	defer func() {
		logger.Info("stopped sending actions to cluster")
	}()

	key := fmt.Sprintf("%s/%s", accountName, clusterName)

	consumer, err := g.createConsumer(server.Context(), accountName, clusterName)
	if err != nil {
		return klErrors.NewE(err)
	}

	// TODO: implement cluster online feature, so that we can mark the cluster as online/offline
	logger.Info("consumer is available now")

	if _, err := g.infraClient.MarkClusterOnlineAt(server.Context(), &infra.MarkClusterOnlineAtIn{
		AccountName: accountName,
		ClusterName: clusterName,
		Timestamp:   timestamppb.New(time.Now()),
	}); err != nil {
		logger.Error("marking cluster online", "err", err)
		return klErrors.NewE(err)
	}

	go func() {
		<-server.Context().Done()
		logger.Debug("server context has been closed")
		delete(g.consumers, key)
		if err := consumer.Stop(context.TODO()); err != nil {
			logger.Error("while stopping consumer", "err", err)
		}
		logger.Info("consumer is closed now")
	}()

	if err := consumer.Consume(func(msg *types.ConsumeMsg) error {
		logger.Info("read message from consumer", "subject", msg.Subject)
		defer func() {
			logger.Info("dispatched message to agent", "subject", msg.Subject)
		}()
		return server.Send(&messages.Action{Message: msg.Payload})
	}, types.ConsumeOpts{
		OnError: func(error) error {
			logger.Info("error occurrred on agent side, while parsing/applying the message, ignoring as we don't want to block the queue")
			return nil
		},
	}); err != nil {
		logger.Error("while consuming messages from consumer", "err", err)
	}

	return nil
}

type ResourceUpdateArgs struct {
	logger          *slog.Logger
	updatesProducer UpdatesProducer

	AccountName string
	ClusterName string

	Message *messages.ResourceUpdate
}

func dispatchResourceUpdate(ctx context.Context, receiver common.MessageReceiver, args ResourceUpdateArgs) (err error) {
	b, err := msgOfficeT.MarshalResourceUpdate(msgOfficeT.ResourceUpdate{
		AccountName:   args.AccountName,
		ClusterName:   args.ClusterName,
		WatcherUpdate: args.Message.Message,
	})
	if err != nil {
		return errors.Wrap(err, "marshalling resource update")
	}

	subject := common.ReceiveFromAgentSubjectName(common.ReceiveFromAgentArgs{
		AccountName: args.AccountName,
		ClusterName: args.ClusterName,
		GVK:         args.Message.Gvk,
		Namespace:   args.Message.Namespace,
		Name:        args.Message.Name,
	}, receiver, common.EventResourceUpdate)

	args.logger.Debug("dispatching to", "subject", subject)

	if err := args.updatesProducer.Produce(ctx, types.ProduceMsg{
		Subject: subject,
		Payload: b,
	}); err != nil {
		return errors.Wrap(err, "producing resource update")
	}

	return nil
}

func NewMessageOfficeServer(producer UpdatesProducer, jc *nats.JetstreamClient, ev *env.Env, d domain.Domain, logger *slog.Logger, infraConn InfraGRPCClient) (messages.MessageDispatchServiceServer, error) {
	return &grpcServer{
		UnimplementedMessageDispatchServiceServer: messages.UnimplementedMessageDispatchServiceServer{},
		infraClient:     infra.NewInfraClient(infraConn),
		logger:          logger,
		updatesProducer: producer,
		consumers:       make(map[string]messaging.Consumer),
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
						fmt.Sprintf("%s.>", common.SendToAgentSubjectPrefix(accountName, clusterName)),
						// common.GetTenantClusterMessagingTopic(accountName, clusterName),
					},
				},
			})
		},
	}, nil
}
