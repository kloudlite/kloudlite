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
	HarborWebhookEndpoint   string `env:"HARBOR_WEBHOOK_ENDPOINT" required:"true"`
	HarborApiVersion        string `env:"HARBOR_API_VERSION" required:"true"`
	HarborWebhookAuthz      string `env:"HARBOR_WEBHOOK_AUTHZ" required:"true"`

	DockerSecretName   string `env:"DOCKER_SECRET_NAME" required:"true"`
	ServiceAccountName string `env:"SERVICE_ACCOUNT_NAME" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
