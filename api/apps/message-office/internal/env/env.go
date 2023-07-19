package env

import "github.com/codingconcepts/env"

type Env struct {
	KafkaBrokers      string `env:"KAFKA_BROKERS"       required:"true"`
	KafkaSaslUsername string `env:"KAFKA_SASL_USERNAME" required:"true"`
	KafkaSaslPassword string `env:"KAFKA_SASL_PASSWORD" required:"true"`

	// for consumers
	KafkaConsumerGroup       string `env:"KAFKA_CONSUMER_GROUP"            required:"true"`
	KafkaTopicStatusUpdates  string `env:"KAFKA_TOPIC_STATUS_UPDATES"      required:"true"`
	KafkaTopicInfraUpdates   string `env:"KAFKA_TOPIC_INFRA_UPDATES"       required:"true"`
	KafkaTopicErrorOnApply   string `env:"KAFKA_TOPIC_ERROR_ON_APPLY"      required:"true"`
	KafkaTopicClusterUpdates string `env:"KAFKA_TOPIC_CLUSTER_UPDATES" required:"true"`

	DbName string `env:"DB_NAME" required:"true"`
	DbUri  string `env:"DB_URI"  required:"true"`

	GrpcPort uint16 `env:"GRPC_PORT" required:"true"`
	HttpPort uint16 `env:"HTTP_PORT" required:"true"`

	// GrpcValidityHeader string `env:"GRPC_VALIDITY_HEADER" required:"true"`
	VectorGrpcAddr string `env:"VECTOR_GRPC_ADDR" required:"true"`
}

func LoadEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
