package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-zookeeper/internal/controllers"
	"operators.kloudlite.io/operators/msvc-zookeeper/internal/env"
)

func main() {
	op := operator.New("zookeeper")

	ev := env.GetEnvOrDie()

	op.RegisterControllers(&controllers.ServiceReconciler{Name: "zookeeper-svc", Env: ev})
	op.Start()
}
