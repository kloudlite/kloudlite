package main

import (
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-n-mres/internal/env"
	"operators.kloudlite.io/operators/msvc-n-mres/internal/mres"
	"operators.kloudlite.io/operators/msvc-n-mres/internal/msvc"
)

func main() {
	op := operator.New("msvc-and-mres")
	ev := env.GetEnvOrDie()
	op.RegisterControllers(
		&msvc.ManagedServiceReconciler{Name: "msvc", Env: ev},
		&mres.ManagedResourceReconciler{Name: "mres", Env: ev},
	)
	op.Start()
}
