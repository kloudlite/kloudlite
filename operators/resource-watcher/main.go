package main

import (
	"context"
	"flag"
	"log"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	// byocClientWatcher "github.com/kloudlite/operator/operators/resource-watcher/internal/controllers/byoc-client"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	serverlessv1 "github.com/kloudlite/operator/apis/serverless/v1"
	"github.com/kloudlite/operator/operator"
	watchAndUpdate "github.com/kloudlite/operator/operators/resource-watcher/internal/controllers/watch-and-update"
	watch_and_update "github.com/kloudlite/operator/operators/resource-watcher/internal/controllers/watch-and-update"
	env "github.com/kloudlite/operator/operators/resource-watcher/internal/env"
	libGrpc "github.com/kloudlite/operator/pkg/grpc"
	"github.com/kloudlite/operator/pkg/logging"
)

func main() {
	var runningOnPlatform bool
	flag.BoolVar(&runningOnPlatform, "running-on-platform", false, "--running-on-platform")

	mgr := operator.New("resource-watcher")

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
	)

	var msgSender watch_and_update.MessageSender
	var err error

	logger := logging.NewOrDie(&logging.Options{Name: "resource-watcher", Dev: mgr.Operator().IsDev})

	if runningOnPlatform {
		msgSender, err = watch_and_update.NewKafkaMessageSender(context.TODO(), ev, logger)
		if err != nil {
			panic(err)
		}
	} else {
		go func() {
			for {
				var err error

				cc, err := libGrpc.ConnectSecure(ev.GrpcAddr)
				if err != nil {
					log.Fatalf("Failed to connect after retries: %v", err)
				}

				msgSender, err = watch_and_update.NewGRPCMessageSender(context.TODO(), cc, ev, logger)
				if err != nil {
					log.Fatalf("Failed to create grpc message sender: %v", err)
				}
			}
		}()
	}

	mgr.RegisterControllers(
		&watchAndUpdate.Reconciler{
			Name:      "watch-and-update",
			Env:       ev,
			MsgSender: msgSender,
		},
		// &byocClientWatcher.Reconciler{
		// 	Name: "byoc-client-watcher",
		// 	Env:  ev,
		// },
	)

	mgr.Start()
}
