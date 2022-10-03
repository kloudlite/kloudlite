package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-n-mres/internal/mres"
	"operators.kloudlite.io/operators/msvc-n-mres/internal/msvc"
)

func main() {
	op := operator.New("msvc-and-mres")
	op.RegisterControllers(
		&msvc.ManagedServiceReconciler{Name: "msvc"},
		&mres.ManagedResourceReconciler{Name: "mres"},
	)
	op.Start()
}
