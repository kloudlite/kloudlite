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
	HttpPort     int    `env:"PORT", required:"true"`
	KafkaBrokers string `env:"BOOTSTRAP_SERVERS", required:"true"`
	isProd       bool   `env:"PROD"`
}

var Module = fx.Module("framework",
	// Load Env
	fx.Provide(func() (*Env, error) {
		var envC Env
		err := config.LoadConfigFromEnv(&envC)
		return &envC, err
	}),
	// Setup Logger
	fx.Provide(func(env *Env) logger.Logger {
		return logger.NewLogger(env.isProd)
	}),
	// Create Producer
	fx.Provide(func(e *Env) (messaging.Producer[messaging.Json], error) {
		return messaging.NewKafkaProducer[messaging.Json](e.KafkaBrokers)
	}),
	// Create Server
	fx.Provide(fiber_app.NewFiberApp),
	// Load App
	app.Module,
	// Start Server with loaded app
	fx.Invoke(func(server *fiber.App, c *Env, lifecycle fx.Lifecycle) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return server.Listen(fmt.Sprintf(":%d", c.HttpPort))
			},
		})
	}),
)
