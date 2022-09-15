package main

import (
	"operators.kloudlite.io/operators/msvc-and-mres/internal/mres"
	"operators.kloudlite.io/operators/msvc-and-mres/internal/msvc"
	"operators.kloudlite.io/operators/operator"
)

func main() {
	op := operator.New("msvc-and-mres")
	op.RegisterControllers(
		&msvc.ManagedServiceReconciler{Name: "msvc"},
		&mres.ManagedResourceReconciler{Name: "mres"},
	)
	op.Start()
}
