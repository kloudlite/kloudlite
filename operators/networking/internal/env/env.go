package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int    `env:"MAX_CONCURRENT_RECONCILES" required:"true"`
	GatewayAdminNamespace   string `env:"GATEWAY_ADMIN_NAMESPACE" default:"kl-gateway"`

	ImageWebhookServer            string `env:"IMAGE_WEBHOOK_SERVER" required:"true"`
	ImageIPManager                string `env:"IMAGE_IP_MANAGER" required:"true"`
	ImageIPBindingController string `env:"IMAGE_IP_BINDING_CONTROLLER" required:"true"`
	ImageDNS                      string `env:"IMAGE_DNS" required:"true"`
	ImageLogsProxy                string `env:"IMAGE_LOGS_PROXY" required:"true"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}

func GetEnvOrDie() *Env {
	ev, err := LoadEnv()
	if err != nil {
		panic(err)
	}
	return ev
}
