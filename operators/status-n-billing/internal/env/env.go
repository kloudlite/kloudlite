package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	KafkaBrokers           string `env:"KAFKA_BROKERS" required:"true"`
	KafkaSASLUsername      string `env:"KAFKA_SASL_USERNAME" required:"true"`
	KafkaSASLPassword      string `env:"KAFKA_SASL_PASSWORD" required:"true"`
	ClusterId              string `env:"CLUSTER_ID" required:"true"`
	KafkaStatusReplyTopic  string `env:"KAFKA_STATUS_REPLY_TOPIC" required:"true"`
	KafkaBillingReplyTopic string `env:"KAFKA_BILLING_REPLY_TOPIC" required:"true"`
}

func GetEnvOrDie() *Env {
	ev := Env{}
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
