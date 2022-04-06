package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	errors "kloudlite.io/pkg/lib-errors"
)

type Json map[string]any

type Producer[T any] interface {
	SendMessage(topic string, key string, message T) error
}

type producer[T any] struct {
	kafkaProducer *kafka.Producer
}

func (m producer[T]) SendMessage(topic string, key string, message T) error {
	msgBody, e := json.Marshal(message)
	errors.AssertNoError(e, fmt.Errorf("failed to marshal message"))
	return m.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: msgBody,
	}, nil)
}

func NewKafkaProducer[T any](kafkaCli KafkaClient) (messenger Producer[T], e error) {
	defer errors.HandleErr(&e)
	p, e := kafka.NewProducer(
		&kafka.ConfigMap{
			"bootstrap.servers": kafkaCli.GetBrokers(),
		},
	)
	errors.AssertNoError(e, fmt.Errorf("failed to create kafka producer"))
	return &producer[T]{
		kafkaProducer: p,
	}, e
}
