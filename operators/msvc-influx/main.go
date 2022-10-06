package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-influx/internal/controllers/bucket"
	"operators.kloudlite.io/operators/msvc-influx/internal/controllers/service"
	"operators.kloudlite.io/operators/msvc-influx/internal/env"
)

func main() {
	mgr := operator.New("influxdb")
	ev := env.GetEnvOrDie()
	mgr.RegisterControllers(
		&service.Reconciler{Name: "service", Env: ev},
		&bucket.Reconciler{Name: "bucket", Env: ev},
	)
	mgr.Start()
}
