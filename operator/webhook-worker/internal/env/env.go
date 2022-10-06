package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	KafkaSASLUsername string `env:"KAFKA_SASL_USERNAME" required:"true"`
	KafkaSASLPassword string `env:"KAFKA_SASL_PASSWORD" required:"true"`

	KafkaBrokers                    string `env:"KAFKA_BROKERS" required:"true"`
	KafkaConsumerGroupId            string `env:"KAFKA_CONSUMER_GROUP_ID" required:"true"`
	KafkaHarborWebhookIncomingTopic string `env:"KAFKA_HARBOR_WEBHOOK_INCOMING_TOPIC" required:""`
	// KafkaWebhookErrorTopic          string `env:"KAFKA_WEBHOOK_ERROR_TOPIC" required:"true"`
}

func GetEnvOrDie() *Env {
	var envVars Env
	if err := env.Set(&envVars); err != nil {
		panic(err)
	}
	return &envVars
}
