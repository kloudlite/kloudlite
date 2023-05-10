package main

import (
	"google.golang.org/grpc"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	serverlessv1 "github.com/kloudlite/operator/apis/serverless/v1"
	"github.com/kloudlite/operator/operator"
	byocClientWatcher "github.com/kloudlite/operator/operators/resource-watcher/internal/controllers/byoc-client"
	watchAndUpdate "github.com/kloudlite/operator/operators/resource-watcher/internal/controllers/watch-and-update"
	env "github.com/kloudlite/operator/operators/resource-watcher/internal/env"
	libGrpc "github.com/kloudlite/operator/pkg/grpc"
)

func main() {
	ev := env.GetEnvOrDie()

	mgr := operator.New("resource-watcher")

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
		clustersv1.AddToScheme,
	)

	mgr.RegisterControllers(
		&watchAndUpdate.Reconciler{
			Name:              "watch-and-update",
			Env:               ev,
			GetGrpcConnection: getGrpcConnection,
		},
		&byocClientWatcher.Reconciler{
			Name: "byoc-client-watcher",
			Env:  ev,
		},
	)
	mgr.Start()
}
