package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	// KafkaSASLUser      string                 `env:"KAFKA_SASL_USER" required:"true"`
	// KafkaSASLPassword  string                 `env:"KAFKA_SASL_PASSWORD" required:"true"`
	// KafkaSASLMechanism redpanda.SASLMechanism `env:"KAFKA_SASL_MECHANISM"`

	// KafkaBrokers           string `env:"KAFKA_BROKERS" required:"true"`
	// KafkaConsumerGroupId   string `env:"KAFKA_CONSUMER_GROUP_ID" required:"true"`
	// KafkaIncomingTopic     string `env:"KAFKA_INCOMING_TOPIC" required:"true"`
	// KafkaErrorOnApplyTopic string `env:"KAFKA_ERROR_ON_APPLY_TOPIC" required:"true"`

	GrpcAddr string `env:"GRPC_ADDR" required:"true"`

	ClusterToken               string `env:"CLUSTER_TOKEN" required:"false"`
	AccessToken                string `env:"ACCESS_TOKEN" required:"false"`
	AccessTokenSecretName      string `env:"ACCESS_TOKEN_SECRET_NAME" required:"true"`
	AccessTokenSecretNamespace string `env:"ACCESS_TOKEN_SECRET_NAMESPACE" required:"true"`

	ClusterName string `env:"CLUSTER_NAME" required:"true"`
	AccountName string `env:"ACCOUNT_NAME" required:"true"`

	HarborSecretName      string `env:"HARBOR_SECRET_NAME" required:"true"`
	HarborSecretNamespace string `env:"HARBOR_SECRET_NAMESPACE" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
