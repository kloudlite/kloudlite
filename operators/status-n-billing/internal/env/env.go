package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD" required:"true"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES" required:"true"`

	KafkaBrokers      string `env:"KAFKA_BROKERS" required:"true"`
	KafkaSASLUsername string `env:"KAFKA_SASL_USERNAME" required:"true"`
	KafkaSASLPassword string `env:"KAFKA_SASL_PASSWORD" required:"true"`
	ClusterId         string `env:"CLUSTER_ID" required:"true"`

	KafkaTopicStatusUpdates      string `env:"KAFKA_TOPIC_STATUS_UPDATES" required:"true"`
	KafkaTopicBillingUpdates     string `env:"KAFKA_TOPIC_BILLING_UPDATES" required:"true"`
	KafkaTopicPipelineRunUpdates string `env:"KAFKA_TOPIC_PIPELINE_RUN_UPDATES" required:"true"`
}

func GetEnvOrDie() *Env {
	ev := Env{}
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
