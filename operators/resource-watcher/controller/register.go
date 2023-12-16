package controller

import (
	"context"
	"log"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
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

func RegisterInto(mgr operator.Operator, runningOnTenant bool) {
	ev, err2 := func() (*env.Env, error) {
		var ev env.Env

		ce, err := env.GetCommonEnv()
		if err != nil {
			return nil, err
		}

		ev.CommonEnv = ce

		if runningOnTenant {
			te, err := env.GetTargetClusterEnvs()
			if err != nil {
				return nil, err
			}
			ev.RunningOnTenantClusterEnv = te
		} else {
			pe, err := env.GetPlatofmrClusterEnvs()
			if err != nil {
				return nil, err
			}
			ev.RunningOnPlatformEnv = pe
		}

		return &ev, nil
	}()

	if err2 != nil {
		panic(err2)
	}

	mgr.AddToSchemes(
		crdsv1.AddToScheme,
		mongodbMsvcv1.AddToScheme, mysqlMsvcv1.AddToScheme, redisMsvcv1.AddToScheme,
		serverlessv1.AddToScheme,
		clustersv1.AddToScheme,
		wireguardv1.AddToScheme,
	)

	logger := logging.NewOrDie(&logging.Options{Name: "resource-watcher", Dev: mgr.Operator().IsDev})

	errCh := make(chan error)

	watchAndUpdateReconciler := &watchAndUpdate.Reconciler{
		Name: "resource-watcher",
		Env:  ev,
	}

	go func() {
		for {
			logger.Infof("connecting to grpc addr: %s", ev.GrpcAddr)

			cc, err := libGrpc.Connect(ev.GrpcAddr, libGrpc.ConnectOpts{
				SecureConnect: runningOnTenant,
				Timeout:       100 * time.Second,
			})
			if err != nil {
				logger.Infof("failed to connect to grpc addr: %s, will be retrying in %d seconds", ev.GrpcAddr, 2)
				<-time.After(2 * time.Second)
				continue
			}

			logger.Infof("grpc connection successfull")

			msgSender, err := watchAndUpdate.NewGRPCMessageSender(context.TODO(), cc, ev, logger, runningOnTenant)
			if err != nil {
				logger.Infof("Failed to create grpc message sender: %v", err)
				<-time.After(2 * time.Second)
				continue
			}

			watchAndUpdateReconciler.MsgSender = msgSender

			connState := cc.GetState()
			for connState != connectivity.Ready && connState != connectivity.Shutdown {
				log.Printf("Connection lost, trying to reconnect")
				errCh <- err
			}
			<-errCh
			cc.Close()
		}
	}()

	mgr.RegisterControllers(
		watchAndUpdateReconciler,
	)
}
