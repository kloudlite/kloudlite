package main

import (
	"fmt"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/controllers/app"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/env"
)

func main() {
	ev := env.GetEnvOrDie()

	fmt.Println("asdfasdf")
	mgr := operator.New("app-n-lambda")
	mgr.AddToSchemes(crdsv1.AddToScheme)
	mgr.RegisterControllers(
		&app.Reconciler{Name: "app", Env: ev},
		// &lambda.Reconciler{Name: "lambda", Env: ev},
	)
	mgr.Start()
}
