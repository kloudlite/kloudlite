package controller

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/nxtcoder17/go.pkgs/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"

	// mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	// redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	// serverlessv1 "github.com/kloudlite/operator/apis/serverless/v1"
	// wireguardv1 "github.com/kloudlite/operator/apis/wireguard/v1"

	watchAndUpdate "github.com/kloudlite/operator/operators/resource-watcher/internal/controllers/watch-and-update"
	env "github.com/kloudlite/operator/operators/resource-watcher/internal/env"
	libGrpc "github.com/kloudlite/operator/pkg/grpc"
	"github.com/kloudlite/operator/toolkit/operator"
)

func RegisterInto(mgr operator.Operator) {
	ev, err := env.GetEnv()
	if err != nil {
		panic(err)
	}

	ev.IsDev = mgr.Operator().IsDev

	mgr.AddToSchemes(
		crdsv1.AddToScheme,
		// mongodbMsvcv1.AddToScheme,
		// mysqlMsvcv1.AddToScheme, redisMsvcv1.AddToScheme,
		// serverlessv1.AddToScheme,
		clustersv1.AddToScheme,
		// wireguardv1.AddToScheme,
	)

	watchAndUpdateReconciler := &watchAndUpdate.Reconciler{
		Env:       ev,
		MsgSender: nil,
	}

	mgr.RegisterControllers(watchAndUpdateReconciler)

	ping := func(cc *grpc.ClientConn) error {
		ctx, cf := context.WithTimeout(context.TODO(), 500*time.Millisecond)
		defer cf()
		msgDispatchCli := messages.NewMessageDispatchServiceClient(cc)
		_, err := msgDispatchCli.Ping(ctx, &messages.Empty{})
		if err != nil {
			log.DefaultLogger().Warn("ping failed, client is disconnected")
			return err
		}
		log.DefaultLogger().Debug("ping is successfull, client is connected")
		return nil
	}

	connectGrpc := func(logger *slog.Logger) error {
		logger.Info("connecting to", "addr", ev.GrpcAddr)

		cc, err := libGrpc.Connect(ev.GrpcAddr, libGrpc.ConnectOpts{
			SecureConnect: ev.GrpcSecureConnect,
			Timeout:       5 * time.Second,
		})
		if err != nil {
			logger.Error("failed to connect via grpc", "addr", ev.GrpcAddr)
			return err
		}

		for {
			connState := cc.GetState()
			logger.Info(fmt.Sprintf("waiting for connection to become %s, current: %s", connectivity.Ready, connState.String()))
			if connState == connectivity.Ready {
				logger.Info("Connected to GRPC server")
				break
			}
			<-time.After(2 * time.Second)
		}

		if err := ping(cc); err != nil {
			return err
		}

		logger.Info("successfully connected to grpc server at", "addr", ev.GrpcAddr)

		ctx, cf := context.WithTimeout(context.TODO(), 2*time.Second)
		defer cf()

		msgSender, err := watchAndUpdate.NewGRPCMessageSender(ctx, cc, ev, logger)
		if err != nil {
			logger.Error("Failed to create grpc message sender, got", "err", err)
			return err
		}

		watchAndUpdateReconciler.MsgSender = msgSender

		defer func() {
			cc.Close()
			watchAndUpdateReconciler.MsgSender = nil
			logger.Info("closed grpc connection")
		}()

		for {
			if err := ping(cc); err != nil {
				return err
			}
			<-time.After(2 * time.Second)
		}
	}

	go func() {
		logger := log.DefaultLogger().With("component", "grpc-client").Slog()
		for {
			connectGrpc(logger)
			<-time.After(2 * time.Second)
		}
	}()
}
