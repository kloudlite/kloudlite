package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"kloudlite.io/pkg/errors"
)

type Json map[string]interface{}

type Producer interface {
	SendMessage(topic string, key string, message Json) error
}

type producer struct {
	kafkaProducer *kafka.Producer
}

func (m producer) SendMessage(topic string, key string, message Json) error {
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

func NewKafkaProducer(kafkaBorkers string) (messenger Producer, e error) {
	defer errors.HandleErr(&e)
	p, e := kafka.NewProducer(
		&kafka.ConfigMap{
			"bootstrap.servers": kafkaBorkers,
		},
	)
	errors.AssertNoError(e, fmt.Errorf("failed to create kafka producer"))
	return &producer{
		kafkaProducer: p,
	}, e
}
