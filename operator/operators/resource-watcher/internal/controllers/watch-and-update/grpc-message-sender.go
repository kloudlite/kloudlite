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
	logger                 logging.Logger
	accessToken            string
	msgDispatchCli         messages.MessageDispatchServiceClient
	messageProtocolVersion string
}

func (g *grpcMsgSender) DispatchContainerRegistryResourceUpdates(ctx MessageSenderContext, ru t.ResourceUpdate) error {
	b, err := json.Marshal(ru)
	if err != nil {
		return err
	}

	dctx, cf := context.WithTimeout(ctx, 1*time.Second)
	defer cf()

	gctx := metadata.NewOutgoingContext(dctx, metadata.Pairs("authorization", g.accessToken))

	errCh := make(chan error, 1)
	execCh := make(chan struct{}, 1)
	go func() {
		if _, err := g.msgDispatchCli.ReceiveContainerRegistryUpdate(gctx, &messages.ResourceUpdate{
			ProtocolVersion: g.messageProtocolVersion,
			Message:         b,
			Gvk:             ru.Object.GetObjectKind().GroupVersionKind().String(),
			Namespace:       ru.Object.GetNamespace(),
			Name:            ru.Object.GetName(),
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
		ctx.logger.Infof("dispatched container registry resource update to message office api")
		return nil
	case err := <-errCh:
		return err
	}
}

// DispatchInfraResourceUpdates implements MessageSender.
func (g *grpcMsgSender) DispatchInfraResourceUpdates(ctx MessageSenderContext, ru t.ResourceUpdate) error {
	b, err := json.Marshal(ru)
	if err != nil {
		return err
	}

	dctx, cf := context.WithTimeout(ctx, 1*time.Second)
	defer cf()

	gctx := metadata.NewOutgoingContext(dctx, metadata.Pairs("authorization", g.accessToken))

	errCh := make(chan error, 1)
	execCh := make(chan struct{}, 1)
	go func() {
		if _, err := g.msgDispatchCli.ReceiveInfraResourceUpdate(gctx, &messages.ResourceUpdate{
			ProtocolVersion: g.messageProtocolVersion,
			Message:         b,
			Gvk:             ru.Object.GetObjectKind().GroupVersionKind().String(),
			Namespace:       ru.Object.GetNamespace(),
			Name:            ru.Object.GetName(),
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
		ctx.logger.WithKV("timestamp", time.Now()).Infof("dispatched infra resource update to message office api")
		return nil
	case err := <-errCh:
		return err
	}
}

// DispatchConsoleResourceUpdates implements MessageSender.
func (g *grpcMsgSender) DispatchConsoleResourceUpdates(ctx MessageSenderContext, ru t.ResourceUpdate) error {
	b, err := json.Marshal(ru)
	if err != nil {
		return err
	}

	dctx, cf := context.WithTimeout(ctx, 1*time.Second)
	defer cf()

	gctx := metadata.NewOutgoingContext(dctx, metadata.Pairs("authorization", g.accessToken))

	errCh := make(chan error, 1)
	execCh := make(chan struct{}, 1)

	go func() {
		if _, err = g.msgDispatchCli.ReceiveConsoleResourceUpdate(gctx, &messages.ResourceUpdate{
			ProtocolVersion: g.messageProtocolVersion,
			Message:         b,
			Gvk:             ru.Object.GetObjectKind().GroupVersionKind().String(),
			Namespace:       ru.Object.GetNamespace(),
			Name:            ru.Object.GetName(),
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
		ctx.logger.Infof("dispatched console resource update to message office api")
		return nil
	case err := <-errCh:
		return err
	}
}

// DispatchIotConsoleResourceUpdates implements MessageSender.
func (g *grpcMsgSender) DispatchIotConsoleResourceUpdates(ctx MessageSenderContext, ru t.ResourceUpdate) error {
	b, err := json.Marshal(ru)
	if err != nil {
		return err
	}

	dctx, cf := context.WithTimeout(ctx, 1*time.Second)
	defer cf()

	gctx := metadata.NewOutgoingContext(dctx, metadata.Pairs("authorization", g.accessToken))

	errCh := make(chan error, 1)
	execCh := make(chan struct{}, 1)

	go func() {
		if _, err = g.msgDispatchCli.ReceiveIotConsoleResourceUpdate(gctx, &messages.ResourceUpdate{
			ProtocolVersion: g.messageProtocolVersion,
			Message:         b,
			Gvk:             ru.Object.GetObjectKind().GroupVersionKind().String(),
			Namespace:       ru.Object.GetNamespace(),
			Name:            ru.Object.GetName(),
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
		ctx.logger.Infof("dispatched iot console resource update to message office api")
		return nil
	case err := <-errCh:
		return err
	}
}

func NewGRPCMessageSender(ctx context.Context, cc *grpc.ClientConn, ev *env.Env, logger logging.Logger) (MessageSender, error) {
	msgDispatchCli := messages.NewMessageDispatchServiceClient(cc)

	authzGrpcCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", ev.AccessToken))

	validationOut, err := msgDispatchCli.ValidateAccessToken(authzGrpcCtx, &messages.ValidateAccessTokenIn{
		ProtocolVersion: ev.GrpcMessagesVersion,
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
		logger:                 logger,
		msgDispatchCli:         msgDispatchCli,
		accessToken:            ev.AccessToken,
		messageProtocolVersion: ev.GrpcMessagesVersion,
	}, nil
}
