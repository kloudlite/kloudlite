package main

import (
	"operators.kloudlite.io/operators/msvc.elasticsearch/internal/controllers"
	"operators.kloudlite.io/operators/operator"
)

func main() {
	op := operator.New("msvc.elasticsearch")
	op.RegisterControllers(&controllers.ServiceReconciler{Name: "service"})
	op.Start()
}
