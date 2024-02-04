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
)

type grpcMsgSender struct {
	logger         logging.Logger
	accountName    string
	clusterName    string
	accessToken    string
	msgDispatchCli messages.MessageDispatchServiceClient
}

func (g *grpcMsgSender) DispatchContainerRegistryResourceUpdates(ctx context.Context, stu t.ResourceUpdate) error {
	b, err := json.Marshal(stu)
	if err != nil {
		return err
	}

	dctx, cf := context.WithTimeout(ctx, 1*time.Second)
	defer cf()

	errCh := make(chan error, 1)
	execCh := make(chan struct{}, 1)
	go func() {
		if _, err := g.msgDispatchCli.ReceiveContainerRegistryUpdate(dctx, &messages.ResourceUpdate{
			AccountName: g.accountName,
			ClusterName: g.clusterName,
			AccessToken: g.accessToken,
			Message:     b,
		}); err != nil {
			errCh <- err
			return
		}
		execCh <- struct{}{}
	}()

	select {
	case <-dctx.Done():
		return dctx.Err()
	case <-execCh:
		g.logger.WithKV("timestamp", time.Now()).Infof("dispatched container registry resource update to message office api")
		return nil
	case err := <-errCh:
		return err
	}
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
		if _, err := g.msgDispatchCli.ReceiveInfraResourceUpdate(ctx, &messages.ResourceUpdate{
			AccountName: g.accountName,
			ClusterName: g.clusterName,
			AccessToken: g.accessToken,
			Message:     b,
		}); err != nil {
			errCh <- err
			return
		}
		execCh <- struct{}{}
	}()

	select {
	case <-dctx.Done():
		return dctx.Err()
	case <-execCh:
		g.logger.WithKV("timestamp", time.Now()).Infof("dispatched infra resource update to message office api")
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
		if _, err = g.msgDispatchCli.ReceiveConsoleResourceUpdate(ctx, &messages.ResourceUpdate{
			AccountName: g.accountName,
			ClusterName: g.clusterName,
			AccessToken: g.accessToken,
			Message:     b,
		}); err != nil {
			errCh <- err
			return
		}
		execCh <- struct{}{}
	}()

	select {
	case <-dctx.Done():
		return dctx.Err()
	case <-execCh:
		g.logger.WithKV("timestamp", time.Now()).Infof("dispatched console resource update to message office api")
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

	return &grpcMsgSender{
		logger:         logger,
		msgDispatchCli: msgDispatchCli,
		accountName:    ev.AccountName,
		clusterName:    ev.ClusterName,
		accessToken:    ev.AccessToken,
	}, nil
}
