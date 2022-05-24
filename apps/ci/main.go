package main

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/ci/internal/framework"
)

func main() {
	module := framework.Module
	fx.New(module).Run()
}
