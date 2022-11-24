package main

import (
	zookeeperMsvcv1 "operators.kloudlite.io/apis/zookeeper.msvc/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-zookeeper/internal/controllers"
	"operators.kloudlite.io/operators/msvc-zookeeper/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()
	op := operator.New("zookeeper")
	op.AddToSchemes(zookeeperMsvcv1.AddToScheme)
	op.RegisterControllers(&controllers.ServiceReconciler{Name: "zookeeper-svc", Env: ev})
	op.Start()
}
