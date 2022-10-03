package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc.zookeeper/internal/controllers"
)

func main() {
	op := operator.New("zookeeper")
	op.RegisterControllers(&controllers.ServiceReconciler{Name: "zookeeper-svc"})
	op.Start()
}
