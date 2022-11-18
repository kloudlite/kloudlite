package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`

	CloudflareWildcardDomains string `env:"CLOUDFLARE_WILDCARD_DOMAINS" required:"true"`
	CloudflareEmail           string `env:"CLOUDFLARE_EMAIL" required:"true"`
	CloudflareSecretName      string `env:"CLOUDFLARE_SECRET_NAME" required:"true"`

	AcmeEmail string `env:"ACME_EMAIL" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
