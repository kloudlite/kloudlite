package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`
	// comma separated list of kafka topics
	// DefaultCreateTopics string `env:"DEFAULT_KAFKA_TOPICS" required:"true"`
	// PushAccessToTopics string `env:"PUSH_ACCESS_TO_TOPICS"`

	RedpandaSecretName      string `env:"REDPANDA_SECRET_NAME" required:"true"`
	RedpandaSecretNamespace string `env:"REDPANDA_SECRET_NAMESPACE" required:"true"`
}

// topics: incoming, status-reply, billing-reply
// acl user: kl-01
// kl-01-incoming,

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
