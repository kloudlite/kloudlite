package main

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	serverlessv1 "github.com/kloudlite/operator/apis/serverless/v1"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/kloudlite/operator/operator"
	statusWatcher "github.com/kloudlite/operator/operators/status-n-billing/internal/controllers/status-watcher"
	env "github.com/kloudlite/operator/operators/status-n-billing/internal/env"
	libGrpc "github.com/kloudlite/operator/pkg/grpc"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	"google.golang.org/grpc"
)

type grpcHandler struct {
	inMemCounter int64
	yamlClient   *kubectl.YAMLClient
	logger       logging.Logger
	ev           *env.Env
	// gConn        *grpc.ClientConn
	// var inMemCounter int64 = 0
	errorsCli      messages.MessageDispatchService_ReceiveErrorsClient
	msgDispatchCli messages.MessageDispatchServiceClient
}

func main() {
	ev := env.GetEnvOrDie()

	// g := grpcHandler{
	// 	inMemCounter: 0,
	// 	// yamlClient:   yamlClient,
	// 	// logger:       logger,
	// 	ev: ev,
	// }


	// producer, err := redpanda.NewProducer(
	// 	ev.KafkaBrokers, redpanda.ProducerOpts{
	// 		SASLAuth: &redpanda.KafkaSASLAuth{
	// 			// SASLMechanism: redpanda.ScramSHA256,
	// 			User:     ev.KafkaSASLUsername,
	// 			Password: ev.KafkaSASLPassword,
	// 		},
	// 	},
	// )
	//
	// if err != nil {
	// 	panic(errors.NewEf(err, "creating redpanda producer"))
	// }
	// defer producer.Close()

	// timeout, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancelFn()
	// if err := producer.Ping(timeout); err != nil {
	// 	panic(fmt.Errorf("failed to ping kafka producer as %s", err.Error()))
	// }

	mgr := operator.New("status-n-billing-watcher")

	getGrpcConnection := func() (*grpc.ClientConn, error) {
		if mgr.Operator().IsDev {
			return libGrpc.Connect(ev.GrpcAddr)
		}
		return libGrpc.ConnectSecure(ev.GrpcAddr)
	}

	mgr.AddToSchemes(
		crdsv1.AddToScheme,
		mongodbMsvcv1.AddToScheme, mysqlMsvcv1.AddToScheme, redisMsvcv1.AddToScheme,
		serverlessv1.AddToScheme,
	)

	mgr.RegisterControllers(
		&statusWatcher.Reconciler{
			Name:              "status-watcher",
			Env:               ev,
			GetGrpcConnection: getGrpcConnection,
			// Producer: producer,
			// Notifier: types.NewNotifier(ev.ClusterId, producer, ev.KafkaTopicStatusUpdates),
		},
		//&pipelineRunWatcher.Reconciler{
		//	Name:       "pipeline-run",
		//	Env:        ev,
		//	Producer:   producer,
		//	KafkaTopic: ev.KafkaTopicPipelineRunUpdates,
		//},
		//&billingWatcher.Reconciler{
		//	Name:     "billing-watcher",
		//	Env:      ev,
		//	Notifier: types.NewNotifier(ev.ClusterId, producer, ev.KafkaTopicBillingUpdates),
		//},
	)
	mgr.Start()
}
