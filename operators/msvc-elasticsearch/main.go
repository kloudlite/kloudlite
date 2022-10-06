package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-elasticsearch/internal/controllers"
	"operators.kloudlite.io/operators/msvc-elasticsearch/internal/env"
)

func main() {
	op := operator.New("msvc-elasticsearch")
	ev := env.GetEnvOrDie()
	op.RegisterControllers(&controllers.ServiceReconciler{Name: "service", Env: ev})
	op.Start()
}
