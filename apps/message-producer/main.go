package main

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/message-producer/internal/framework"
	"kloudlite.io/pkg/config"
)

func main() {
	config.LoadEnv()
	fx.New(framework.Module).Run()
}
