package controller

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	t "github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/errors"
	lib_grpc "github.com/kloudlite/operator/pkg/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
)

type grpcMsgSender struct {
	logger                 *slog.Logger
	accessToken            string
	msgDispatchCli         messages.MessageDispatchServiceClient
	messageProtocolVersion string
}

// DispatchResourceUpdates implements MessageSender.
func (g *grpcMsgSender) DispatchResourceUpdates(ctx Context, ru t.ResourceUpdate) error {
	b, err := json.Marshal(ru)
	if err != nil {
		return err
	}

	dctx, cf := context.WithTimeout(ctx, 1*time.Second)
	defer cf()

	gctx := metadata.NewOutgoingContext(dctx, metadata.Pairs("authorization", g.accessToken))

	if _, err = g.msgDispatchCli.ReceiveResourceUpdate(gctx, &messages.ResourceUpdate{
		ProtocolVersion: g.messageProtocolVersion,
		Message:         b,
		Gvk:             ru.Object.GetObjectKind().GroupVersionKind().String(),
		Namespace:       ru.Object.GetNamespace(),
		Name:            ru.Object.GetName(),
	}); err != nil {
		return err
	}

	ctx.Logger.Debug("dispatched resource update to message office api")
	return nil
}

func NewGRPCMessageSender(ctx context.Context, ev *Env) (MessageSender, error) {
	cc, err := lib_grpc.Connect(ev.GrpcAddr, lib_grpc.ConnectOpts{
		SecureConnect: ev.GrpcSecureConnect,
		Timeout:       2 * time.Second,
	})
	if err != nil {
		slog.Error("failed to connect at", "grpc-addr", ev.GrpcAddr)
		return nil, err
	}

	for {
		connState := cc.GetState()
		slog.Debug("waiting for connection to become ready", "current.state", connState.String())
		if connState == connectivity.Ready {
			slog.Debug("Connected to GRPC server")
			break
		}
		<-time.After(2 * time.Second)
	}

	slog.Debug("successfully connected to grpc server at", "addr", ev.GrpcAddr)

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
		ctx.Logger.Error(err.Error())
		return nil, err
	}

	return &grpcMsgSender{
		logger:                 slog.Default(),
		msgDispatchCli:         msgDispatchCli,
		accessToken:            ev.AccessToken,
		messageProtocolVersion: ev.GrpcMessagesVersion,
	}, nil
}
