package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc.elasticsearch/internal/controllers"
)

func main() {
	op := operator.New("msvc.elasticsearch")
	op.RegisterControllers(&controllers.ServiceReconciler{Name: "service"})
	op.Start()
}
