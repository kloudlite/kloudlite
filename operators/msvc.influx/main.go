package main

import (
	"operators.kloudlite.io/operators/msvc.influx/internal/controllers/bucket"
	"operators.kloudlite.io/operators/msvc.influx/internal/controllers/service"
	"operators.kloudlite.io/operators/operator"
)

func main() {
	mgr := operator.New("influxdb")
	mgr.RegisterControllers(
		&service.Reconciler{Name: "service"},
		&bucket.Reconciler{Name: "bucket"},
	)
	mgr.Start()
}
