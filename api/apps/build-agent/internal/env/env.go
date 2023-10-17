package env

import "github.com/codingconcepts/env"

type Env struct {
	KafkaBrokers string `env:"KAFKA_BROKERS" required:"true"`

	KafkaBuildTopics   string `env:"KAFKA_BUILD_TOPICS" required:"true"`
	KafkaConsumerGroup string `env:"KAFKA_CONSUMER_GROUP" required:"true"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, err
	}
	return &e, nil
}
