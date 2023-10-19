package env

import "github.com/codingconcepts/env"

type Env struct {
	KafkaBrokers                string `env:"KAFKA_BROKERS" required:"true"`
	WaitQueueKafkaTopic         string `env:"WAIT_QUEUE_KAFKA_TOPIC" required:"true"`
	WaitQueueKafkaConsumerGroup string `env:"WAIT_QUEUE_KAFKA_CONSUMER_GROUP" required:"true"`

	RedpandaHttpAddr string `env:"REDPANDA_HTTP_ADDR" required:"true"`

	NewTopicPartitionsCount  string `env:"NEW_TOPIC_PARTITIONS_COUNT" required:"true"`
	NewTopicReplicationCount string `env:"NEW_TOPIC_REPLICATION_COUNT" required:"true"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
