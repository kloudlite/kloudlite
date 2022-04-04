package main

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/message-producer/internal/framework"
)

func main() {
	fx.New(framework.Module).Run()
}
