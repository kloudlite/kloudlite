package main

import (
	"flag"
	"fmt"

	"go.uber.org/fx"
	"kloudlite.io/apps/auth/internal/framework"
	"kloudlite.io/pkg/logging"
)

func main() {
	fmt.Println("Hello, World!2")
	isDev := flag.Bool("dev", false, "--dev")
	flag.Parse()
	fx.New(
		framework.Module,
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.New(&logging.Options{Name: "auth", Dev: *isDev})
			},
		),
	).Run()
}
