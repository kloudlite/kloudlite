package controller

import (
	"context"
	"errors"
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

func RegisterInto(mgr operator.Operator, runningOnPlatform bool) {
	ev, err2 := func() (*env.Env, error) {
		ce, err := env.GetCommonEnv()
		if err != nil {
			return nil, err
		}

		if runningOnPlatform {
			pe, err := env.GetPlatformEnv()
			if err != nil {
				return nil, err
			}

			return &env.Env{
				CommonEnv:            ce,
				RunningOnPlatformEnv: pe,
			}, nil
		}

		te, err := env.GetTargetClusterEnvs()
		if err != nil {
			return nil, err
		}
		return &env.Env{
			CommonEnv:                 ce,
			RunningOnTargetClusterEnv: te,
		}, nil
	}()

	if err2 != nil {
		panic(err2)
	}

	if runningOnPlatform {
		if ev.KafkaBrokers == "" || ev.KafkaInfraUpdatesTopic == "" || ev.KafkaResourceUpdatesTopic == "" {
			panic("env-var KAFKA_BROKERS, KAFKA_INFRA_UPDATES_TOPIC, KAFKA_RESOURCE_UPDATES_TOPIC are required, when running with --running-on-platform")
		}
	} else {
		if ev.GrpcAddr == "" {
			panic("env-var GRPC_ADDR is required, when running on target clusters")
		}
	}

	mgr.AddToSchemes(
		crdsv1.AddToScheme,
		mongodbMsvcv1.AddToScheme, mysqlMsvcv1.AddToScheme, redisMsvcv1.AddToScheme,
		serverlessv1.AddToScheme,
		clustersv1.AddToScheme,
		wireguardv1.AddToScheme,
	)

	var msgSender watchAndUpdate.MessageSender
	var err error

	logger := logging.NewOrDie(&logging.Options{Name: "resource-watcher", Dev: mgr.Operator().IsDev})

	if runningOnPlatform {
		getMsgSender := func() (watchAndUpdate.MessageSender, error) {
			return watchAndUpdate.NewKafkaMessageSender(context.TODO(), ev, logger)
		}
		msgSender, err = getMsgSender()
		if err != nil {
			if errors.As(err, &watchAndUpdate.ErrConnect{}) {
				go func() {
					for {
						msgSender, err = getMsgSender()
						if err == nil {
							break
						}
						logger.Infof("Failed to connect to kafka, retrying in another 5 seconds")
						<-time.After(5 * time.Second)
					}
				}()
				return
			}
			panic(err)
		}
	} else {
		var err error

		cc, err := libGrpc.ConnectSecure(ev.GrpcAddr)
		if err != nil {
			log.Fatalf("Failed to connect after retries: %v", err)
		}

		msgSender, err = watchAndUpdate.NewGRPCMessageSender(context.TODO(), cc, ev, logger)
		if err != nil {
			log.Fatalf("Failed to create grpc message sender: %v", err)
		}

		go func() {
			errCh := make(chan error)
			for {
				var err error

				cc, err := libGrpc.ConnectSecure(ev.GrpcAddr)
				if err != nil {
					log.Fatalf("Failed to connect after retries: %v", err)
				}

				msgSender, err = watchAndUpdate.NewGRPCMessageSender(context.TODO(), cc, ev, logger)
				if err != nil {
					log.Fatalf("Failed to create grpc message sender: %v", err)
				}

				connState := cc.GetState()
				for connState != connectivity.Ready && connState != connectivity.Shutdown {
					log.Printf("Connection lost, trying to reconnect")
					errCh <- err
				}
				<-errCh
				cc.Close()
			}
		}()
	}

	mgr.RegisterControllers(
		&watchAndUpdate.Reconciler{
			Name:      "resource-watcher",
			Env:       ev,
			MsgSender: msgSender,
		},
	)
}
