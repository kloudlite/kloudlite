package main

import (
	influxDB "operators.kloudlite.io/apis/influxdb.msvc/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-influx/internal/controllers/bucket"
	"operators.kloudlite.io/operators/msvc-influx/internal/controllers/service"
	"operators.kloudlite.io/operators/msvc-influx/internal/env"
)

func main() {
	mgr := operator.New("influxdb")
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(influxDB.AddToScheme)
	mgr.RegisterControllers(
		&service.Reconciler{Name: "service", Env: ev},
		&bucket.Reconciler{Name: "bucket", Env: ev},
	)
	mgr.Start()
}
