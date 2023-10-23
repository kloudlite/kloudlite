package watch_and_update

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/kloudlite/operator/operators/resource-watcher/internal/env"
	t "github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type grpcMsgSender struct {
	errCh               chan error
	resourceMessagesCli messages.MessageDispatchService_ReceiveResourceUpdatesClient
	infraMessagesCli    messages.MessageDispatchService_ReceiveInfraUpdatesClient
	logger              logging.Logger
}

// DispatchInfraUpdates implements MessageSender.
func (g *grpcMsgSender) DispatchInfraUpdates(ctx context.Context, ru t.ResourceUpdate) error {
	b, err := json.Marshal(ru)
	if err != nil {
		return err
	}

	if err = g.infraMessagesCli.Send(&messages.InfraUpdate{Message: b}); err != nil {
		g.errCh <- err
		return err
	}

	g.logger.WithKV("timestamp", time.Now()).Infof("dispatched update to message office api")
	return nil
}

// DispatchResourceUpdates implements MessageSender.
func (g *grpcMsgSender) DispatchResourceUpdates(ctx context.Context, ru t.ResourceUpdate) error {
	b, err := json.Marshal(ru)
	if err != nil {
		return err
	}
	if err = g.resourceMessagesCli.Send(&messages.ResourceUpdate{Message: b}); err != nil {
		g.errCh <- err
		return err
	}

	g.logger.WithKV("timestamp", time.Now()).Infof("dispatched update to message office api")
	return nil
}

func NewGRPCMessageSender(ctx context.Context, cc *grpc.ClientConn, ev *env.Env, logger logging.Logger) (MessageSender, error) {
	msgDispatchCli := messages.NewMessageDispatchServiceClient(cc)

	validationOut, err := msgDispatchCli.ValidateAccessToken(context.TODO(), &messages.ValidateAccessTokenIn{
		AccountName: ev.AccountName,
		ClusterName: ev.ClusterName,
		AccessToken: ev.AccessToken,
	})

	if err != nil || validationOut == nil || !validationOut.Valid {
		logger.Infof("accessToken is invalid, requesting new accessToken ...")
	}

	if validationOut != nil && validationOut.Valid {
		logger.Infof("accessToken is valid, proceeding with it ...")
	}

	outgoingCtx := metadata.NewOutgoingContext(context.TODO(), metadata.Pairs("authorization", ev.AccessToken))

	resourceMessagesCli, err := msgDispatchCli.ReceiveResourceUpdates(outgoingCtx)
	if err != nil {
		logger.Errorf(err, "ReceiveResourceUpdates")
		return nil, err
	}

	infraMessagesCli, err := msgDispatchCli.ReceiveInfraUpdates(outgoingCtx)
	if err != nil {
		logger.Errorf(err, "ReceiveInfraUpdates")
		return nil, err
	}

	return &grpcMsgSender{
		logger:              logger,
		resourceMessagesCli: resourceMessagesCli,
		infraMessagesCli:    infraMessagesCli,
	}, nil
}
