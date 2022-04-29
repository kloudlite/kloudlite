package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/fx"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"kloudlite.io/pkg/errors"
)

type Json map[string]any

type Producer[T any] interface {
	Connect(cxt context.Context) error
	Close(ctx context.Context)
	SendMessage(topic string, key string, message T) error
}

type producer[T any] struct {
	kafkaBrokers  string
	kafkaProducer *kafka.Producer
}

func (m *producer[T]) Connect(cxt context.Context) error {
	p, e := kafka.NewProducer(
		&kafka.ConfigMap{
			"bootstrap.servers": m.kafkaBrokers,
		},
	)
	if e != nil {
		return errors.Wrap(e, "failed to create kafka producer")
	}
	m.kafkaProducer = p
	return nil
}

func (m *producer[T]) Close(ctx context.Context) {
	m.kafkaProducer.Close()
}

func (m *producer[T]) SendMessage(topic string, key string, message T) error {
	msgBody, e := json.Marshal(message)
	if e != nil {
		fmt.Println(e)
		return e
	}
	return m.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic: &topic,
		},
		Key:   []byte(key),
		Value: msgBody,
	}, nil)
}

func NewKafkaProducer[T any](kafkaCli KafkaClient) (messenger Producer[T], e error) {
	return &producer[T]{
		kafkaBrokers: kafkaCli.GetBrokers(),
	}, e
}

func NewFxKafkaProducer[T any]() fx.Option {
	return fx.Module("producer",
		fx.Provide(func(c KafkaClient) (Producer[T], error) {
			return NewKafkaProducer[T](c)
		}),
		fx.Invoke(func(p Producer[T], lifecycle fx.Lifecycle) {
			lifecycle.Append(fx.Hook{
				OnStart: func(context context.Context) error {
					return p.Connect(context)
				},
				OnStop: func(context context.Context) error {
					p.Close(context)
					return nil
				},
			})
		}),
	)
}
