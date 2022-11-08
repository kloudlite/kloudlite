package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`
	AclAllowedOperations    string        `env:"ACL_ALLOWED_OPERATIONS" required:"true"`

	RedpandaDefaultSecretNamespace string `env:"REDPANDA_DEFAULT_SECRET_NAMESPACE" required:"true"`
	RepdandaDefaultSecretName      string `env:"REDPANDA_DEFAULT_SECRET_NAME" required:"true"`

	MustHaveTopics string `env:"MUST_HAVE_TOPICS" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
