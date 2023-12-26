package watch_and_update

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/kloudlite/operator/operators/resource-watcher/internal/env"
	t "github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type grpcMsgSender struct {
	getResourceMessagesCli func() (messages.MessageDispatchService_ReceiveResourceUpdatesClient, error)
	rmc                    messages.MessageDispatchService_ReceiveResourceUpdatesClient
	getInfraMessagesCli    func() (messages.MessageDispatchService_ReceiveInfraUpdatesClient, error)
	imc                    messages.MessageDispatchService_ReceiveInfraUpdatesClient

	logger logging.Logger
}

// DispatchInfraResourceUpdates implements MessageSender.
func (g *grpcMsgSender) DispatchInfraResourceUpdates(ctx context.Context, ru t.ResourceUpdate) error {
	b, err := json.Marshal(ru)
	if err != nil {
		return err
	}

	dctx, cf := context.WithTimeout(ctx, 1*time.Second)
	defer cf()

	errCh := make(chan error, 1)
	execCh := make(chan struct{}, 1)
	go func() {
		if err = g.imc.Send(&messages.InfraUpdate{Message: b}); err != nil {
			// replace streaming client
			g.imc, err = g.getInfraMessagesCli()
			errCh <- err
			return
		}
		execCh <- struct{}{}
	}()

	select {
	case <-dctx.Done():
		return dctx.Err()
	case <-execCh:
		g.logger.WithKV("timestamp", time.Now()).Infof("dispatched update to message office api")
		return nil
	case err := <-errCh:
		return err
	}
}

// DispatchConsoleResourceUpdates implements MessageSender.
func (g *grpcMsgSender) DispatchConsoleResourceUpdates(ctx context.Context, ru t.ResourceUpdate) error {
	b, err := json.Marshal(ru)
	if err != nil {
		return err
	}

	dctx, cf := context.WithTimeout(ctx, 1*time.Second)
	defer cf()

	errCh := make(chan error, 1)
	execCh := make(chan struct{}, 1)

	go func() {
		if err = g.rmc.Send(&messages.ResourceUpdate{Message: b}); err != nil {
			// replace streaming client
			g.rmc, err = g.getResourceMessagesCli()
			errCh <- err
			return
		}
		execCh <- struct{}{}
	}()

	select {
	case <-dctx.Done():
		return dctx.Err()
	case <-execCh:
		g.logger.WithKV("timestamp", time.Now()).Infof("dispatched update to message office api")
		return nil
	case err := <-errCh:
		return err
	}
}

func NewGRPCMessageSender(ctx context.Context, cc *grpc.ClientConn, ev *env.Env, logger logging.Logger) (MessageSender, error) {
	msgDispatchCli := messages.NewMessageDispatchServiceClient(cc)

	validationOut, err := msgDispatchCli.ValidateAccessToken(ctx, &messages.ValidateAccessTokenIn{
		AccountName: ev.AccountName,
		ClusterName: ev.ClusterName,
		AccessToken: ev.AccessToken,
	})
	if err != nil {
		return nil, err
	}

	if validationOut == nil || !validationOut.Valid {
		err := errors.Newf("accessToken is invalid, aborting")
		logger.Error(err)
		return nil, err
	}

	outgoingCtx := func() context.Context {
		return metadata.NewOutgoingContext(context.TODO(), metadata.Pairs("authorization", ev.AccessToken))
	}

	getResourceMessagesCli := func() (messages.MessageDispatchService_ReceiveResourceUpdatesClient, error) {
		return msgDispatchCli.ReceiveResourceUpdates(outgoingCtx())
	}

	getInfraMessagesCli := func() (messages.MessageDispatchService_ReceiveInfraUpdatesClient, error) {
		return msgDispatchCli.ReceiveInfraUpdates(outgoingCtx())
	}

	rmc, err := getResourceMessagesCli()
	if err != nil {
		return nil, err
	}

	imc, err := getInfraMessagesCli()
	if err != nil {
		return nil, err
	}

	return &grpcMsgSender{
		logger:                 logger,
		getResourceMessagesCli: getResourceMessagesCli,
		rmc:                    rmc,
		imc:                    imc,
		getInfraMessagesCli:    getInfraMessagesCli,
	}, nil
}
