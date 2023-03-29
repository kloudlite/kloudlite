package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/operator/pkg/redpanda"
)

type Env struct {
	KafkaSASLUser      string                 `env:"KAFKA_SASL_USER" required:"true"`
	KafkaSASLPassword  string                 `env:"KAFKA_SASL_PASSWORD" required:"true"`
	KafkaSASLMechanism redpanda.SASLMechanism `env:"KAFKA_SASL_MECHANISM"`

	KafkaBrokers           string `env:"KAFKA_BROKERS" required:"true"`
	KafkaConsumerGroupId   string `env:"KAFKA_CONSUMER_GROUP_ID" required:"true"`
	KafkaIncomingTopic     string `env:"KAFKA_INCOMING_TOPIC" required:"true"`
	KafkaErrorOnApplyTopic string `env:"KAFKA_ERROR_ON_APPLY_TOPIC" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
