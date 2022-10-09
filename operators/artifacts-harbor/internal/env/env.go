package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`

	HarborAdminUsername     string `env:"HARBOR_ADMIN_USERNAME" required:"true"`
	HarborAdminPassword     string `env:"HARBOR_ADMIN_PASSWORD" required:"true"`
	HarborImageRegistryHost string `env:"HARBOR_IMAGE_REGISTRY_HOST" required:"true"`
	HarborWebhookAddr       string `env:"HARBOR_WEBHOOK_ADDR" required:"true"`
	HarborApiVersion        string `env:"HARBOR_API_VERSION" required:"false"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
