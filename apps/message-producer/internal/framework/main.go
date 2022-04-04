package framework

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"kloudlite.io/apps/message-producer/internal/app"
	"kloudlite.io/pkg/config"
	fiber_app "kloudlite.io/pkg/fiber-app"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
)

type Env struct {
	config.BaseEnv
	HttpPort     int    `env:"PORT", required:"true"`
	KafkaBrokers string `env:"BOOTSTRAP_SERVERS", required:"true"`
}

var Module = fx.Module("framework",
	fx.Provide(logger.NewLogger),
	fx.Provide(func() *Env {
		var envC *Env
		envC.Load()
		return envC
	}),
	fx.Provide(func(e *Env) (messaging.Producer, error) {
		return messaging.NewKafkaProducer(e.KafkaBrokers)
	}),
	fx.Provide(fiber_app.NewFiberApp),
	app.Module,
	fx.Invoke(func(server *fiber.App, c *Env, lifecycle fx.Lifecycle) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return server.Listen(fmt.Sprintf(":%d", c.HttpPort))
			},
		})
	}),
)
