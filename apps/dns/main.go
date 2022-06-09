package main

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/dns/internal/framework"
)

func main() {
	fx.New(framework.Module).Run()
}
