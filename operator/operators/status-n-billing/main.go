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
