package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc.mysql/internal/controllers/database"
	"operators.kloudlite.io/operators/msvc.mysql/internal/controllers/standalone"
)

func main() {
	mgr := operator.New("msvc.mysql")
	mgr.RegisterControllers(
		&standalone.ServiceReconciler{Name: "standalone-service"},
		&database.Reconciler{Name: "database"},
	)
	mgr.Start()
}
