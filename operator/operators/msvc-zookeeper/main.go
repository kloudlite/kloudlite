package main

import (
	zookeeperMsvcv1 "github.com/kloudlite/operator/apis/zookeeper.msvc/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/msvc-zookeeper/internal/controllers"
	"github.com/kloudlite/operator/operators/msvc-zookeeper/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	op := operator.New("zookeeper")
	op.AddToSchemes(zookeeperMsvcv1.AddToScheme)
	op.RegisterControllers(&controllers.ServiceReconciler{Name: "zookeeper-svc", Env: ev})
	op.Start()
}
