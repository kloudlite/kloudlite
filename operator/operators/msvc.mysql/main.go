package main

import (
	"operators.kloudlite.io/operators/msvc.mysql/internal/controllers/database"
	"operators.kloudlite.io/operators/msvc.mysql/internal/controllers/standalone"
	"operators.kloudlite.io/operators/operator"
)

func main() {
	mgr := operator.New("msvc.mysql")
	mgr.RegisterControllers(
		&standalone.ServiceReconciler{Name: "standalone-service"},
		&database.Reconciler{Name: "database"},
	)
	mgr.Start()
}
