package main

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/framework"
)

func main() {
	fx.New(framework.Module).Run()
}
