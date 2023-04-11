package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD" required:"true"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES" required:"true"`
	// HelmReleaseName         string        `env:"HELM_RELEASE_NAME" required:"true"`
	HelmReleaseNamespace    string        `env:"HELM_RELEASE_NAMESPACE" required:"true"`

	KafkaTopicBYOCClientUpdates string `env:"KAFKA_TOPIC_BYOC_CLIENT_UPDATES" required:"true"`

	KafkaBrokers      string `env:"KAFKA_BROKERS" required:"true"`
	KafkaSASLUsername string `env:"KAFKA_SASL_USERNAME" required:"true"`
	KafkaSASLPassword string `env:"KAFKA_SASL_PASSWORD" required:"true"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
