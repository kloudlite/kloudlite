package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD" required:"true"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES" required:"true"`

	KafkaBrokers           string `env:"KAFKA_BROKERS" required:"true"`
	KafkaSASLUsername      string `env:"KAFKA_SASL_USERNAME" required:"true"`
	KafkaSASLPassword      string `env:"KAFKA_SASL_PASSWORD" required:"true"`
	ClusterId              string `env:"CLUSTER_ID" required:"true"`
	KafkaStatusReplyTopic  string `env:"KAFKA_STATUS_REPLY_TOPIC" required:"true"`
	KafkaBillingReplyTopic string `env:"KAFKA_BILLING_REPLY_TOPIC" required:"true"`
	KafkaPipelineRunTopic  string `env:"KAFKA_PIPELINE_RUN_TOPIC" required:"true"`
}

func GetEnvOrDie() *Env {
	ev := Env{}
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
