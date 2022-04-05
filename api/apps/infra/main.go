package main

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/framework"
)

func main() {
	fx.New(framework.Module).Run()
}
