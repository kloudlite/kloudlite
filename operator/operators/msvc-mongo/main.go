package main

import (
	"operators.kloudlite.io/operators/msvc-mongo/internal/controllers/database"
	"operators.kloudlite.io/operators/msvc-mongo/internal/controllers/standalone"
	"operators.kloudlite.io/operators/operator"
)

func main() {
	op := operator.New("mongodb")
	op.RegisterControllers(
		&standalone.ServiceReconciler{Name: "mongodb-standalone-svc"},
		&database.Reconciler{Name: "mongodb-database"},
	)
	op.Start()
}
