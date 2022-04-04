package main

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/wireguard/internal/framework"
	"kloudlite.io/pkg/config"
)

func main() {
	config.LoadDotEnv()
	fx.New(framework.Module).Run()
}
