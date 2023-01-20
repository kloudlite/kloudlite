package main

import (
	influxDB "github.com/kloudlite/operator/apis/influxdb.msvc/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/msvc-influx/internal/controllers/bucket"
	"github.com/kloudlite/operator/operators/msvc-influx/internal/controllers/service"
	"github.com/kloudlite/operator/operators/msvc-influx/internal/env"
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
