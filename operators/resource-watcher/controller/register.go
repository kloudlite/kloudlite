package controller

import (
	"context"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	serverlessv1 "github.com/kloudlite/operator/apis/serverless/v1"
	wireguardv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operator"
	watchAndUpdate "github.com/kloudlite/operator/operators/resource-watcher/internal/controllers/watch-and-update"
	env "github.com/kloudlite/operator/operators/resource-watcher/internal/env"
	libGrpc "github.com/kloudlite/operator/pkg/grpc"
	"github.com/kloudlite/operator/pkg/logging"
)

func RegisterInto(mgr operator.Operator) {
	ev, err := env.GetEnv()
	if err != nil {
		panic(err)
	}

	mgr.AddToSchemes(
		crdsv1.AddToScheme,
		mongodbMsvcv1.AddToScheme, mysqlMsvcv1.AddToScheme, redisMsvcv1.AddToScheme,
		serverlessv1.AddToScheme,
		clustersv1.AddToScheme,
		wireguardv1.AddToScheme,
	)

	logger := logging.NewOrDie(&logging.Options{Name: "resource-watcher", Dev: mgr.Operator().IsDev})

	watchAndUpdateReconciler := &watchAndUpdate.Reconciler{
		Name:      "resource-watcher",
		Env:       ev,
		MsgSender: nil,
	}

	mgr.RegisterControllers(
		watchAndUpdateReconciler,
	)

	ping := func(cc *grpc.ClientConn) error {
		ctx, cf := context.WithTimeout(context.TODO(), 500*time.Millisecond)
		defer cf()
		msgDispatchCli := messages.NewMessageDispatchServiceClient(cc)
		_, err := msgDispatchCli.Ping(ctx, &messages.Empty{})
		if err != nil {
			logger.Infof("ping failed, client is disconnected")
			return err
		}
		logger.Debugf("ping is successfull, client is connected")
		return nil
	}

	connectGrpc := func(logger logging.Logger) error {
		logger.Infof("connecting to addr: %s", ev.GrpcAddr)

		cc, err := libGrpc.Connect(ev.GrpcAddr, libGrpc.ConnectOpts{
			SecureConnect: ev.GrpcSecureConnect,
			Timeout:       5 * time.Second,
		})
		if err != nil {
			logger.Infof("failed to connect to grpc addr: %s", ev.GrpcAddr)
			return err
		}

		for {
			connState := cc.GetState()
			logger.Infof("waiting for connection to become %s, current: %s", connectivity.Ready, connState.String())
			if connState == connectivity.Ready {
				logger.Infof("Connected to GRPC server")
				break
			}
			<-time.After(2 * time.Second)
		}

		if err := ping(cc); err != nil {
			return err
		}

		logger.Infof("successfully connected to grpc server at %s", ev.GrpcAddr)

		ctx, cf := context.WithTimeout(context.TODO(), 2*time.Second)
		defer cf()
		msgSender, err := watchAndUpdate.NewGRPCMessageSender(ctx, cc, ev, logger)
		if err != nil {
			logger.Infof("Failed to create grpc message sender: %v", err)
			return err
		}

		watchAndUpdateReconciler.MsgSender = msgSender

		defer func() {
			cc.Close()
			watchAndUpdateReconciler.MsgSender = nil
			logger.Infof("closed grpc connection")
		}()

		for {
			if err := ping(cc); err != nil {
				return err
			}
			<-time.After(2 * time.Second)
		}
	}

	go func() {
		logger := logger.WithKV("component", "grpc-client")
		for {
			connectGrpc(logger)
			<-time.After(2 * time.Second)
		}
	}()
}
