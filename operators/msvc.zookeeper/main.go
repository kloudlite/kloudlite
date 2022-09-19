package main

import (
	"operators.kloudlite.io/operators/msvc.zookeeper/internal/controllers"
	"operators.kloudlite.io/operators/operator"
)

func main() {
	op := operator.New("zookeeper")
	op.RegisterControllers(&controllers.ServiceReconciler{Name: "zookeeper-svc"})
	op.Start()
}
