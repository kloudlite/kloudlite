package main

import (
	"operators.kloudlite.io/operators/msvc-redis/internal/controllers/standalone"
	"operators.kloudlite.io/operators/operator"
)

func main() {
	op := operator.New("redis")
	op.RegisterControllers(
		&standalone.ServiceReconciler{Name: "redis-standalone"},
	)
	op.Start()
}
