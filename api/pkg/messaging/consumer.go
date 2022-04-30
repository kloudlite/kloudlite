package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/fx"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logger"
)

type Message []byte

func (m Message) Unmarshal(x any) error {
	return json.Unmarshal(m, &x)
}

type ConsumerCallback func(context context.Context, topic string, message Message) error

type consumer[T KafkaConsumerConfig] struct {
	kafkaConsumer *kafka.Consumer
	topics        []string
	handlers      map[string]func(context context.Context, message Message) error
	stopChan      chan bool
	logger        logger.Logger
}

func (c *consumer[T]) On(topic string, callback func(context context.Context, message Message) error) {
	if c.handlers == nil {
		c.handlers = make(map[string]func(context context.Context, message Message) error)
	}
	c.handlers[topic] = callback
}

func (c *consumer[T]) Unsubscribe(T) error {
	c.stopChan <- true
	return c.kafkaConsumer.Unsubscribe()
}

func (c *consumer[T]) Subscribe(T) error {
	c.stopChan = make(chan bool, 1)
	e := c.kafkaConsumer.SubscribeTopics(c.topics, nil)
	if e != nil {
		return fmt.Errorf("could not subscribe to given topics %v", e)
	}
	go func() {
		var stop = false
		go func() {
			<-c.stopChan
			stop = true
		}()
		for {
			if stop {
				return
			}
			msg, e := c.kafkaConsumer.ReadMessage(-1)
			if e != nil {
				c.logger.Errorf("could not read kafka message because %v", e)
				continue
			}
			topic := *msg.TopicPartition.Topic
			if c.handlers[topic] == nil {
				c.logger.Errorf("no handler for topic %s", topic)
				c.kafkaConsumer.CommitMessage(msg)
				continue
			}
			e = c.handlers[topic](context.Background(), msg.Value)
			if e != nil {
				e = c.handlers[topic](context.Background(), msg.Value)
				if e != nil {
					c.logger.Debug("failed to process message after 2 retries")
				}
			}
			c.kafkaConsumer.CommitMessage(msg)
		}
	}()
	return nil
}

type Consumer[T KafkaConsumerConfig] interface {
	On(topic string, callback func(context context.Context, message Message) error)
	Unsubscribe(T) error
	Subscribe(T) error
}

func NewKafkaConsumer[T KafkaConsumerConfig](
	kafkaCli KafkaClient,
	topics []string,
	consumerGroupId string,
	logger logger.Logger,
) (messenger Consumer[T], e error) {
	defer errors.HandleErr(&e)
	c, e := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  kafkaCli.GetBrokers(),
		"group.id":           consumerGroupId,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	})
	errors.AssertNoError(e, fmt.Errorf("failed to create kafka producer"))
	return &consumer[T]{
		kafkaConsumer: c,
		topics:        topics,
		logger:        logger,
	}, e
}

type KafkaConsumerConfig interface {
	GetSubscriptionTopics() []string
	GetConsumerGroupId() string
}

func NewFxKafkaConsumer[T KafkaConsumerConfig]() fx.Option {
	return fx.Module("consumer",
		fx.Provide(func(c T, kafkaCli KafkaClient, logger logger.Logger) (Consumer[T], error) {
			fmt.Println("subscription", c.GetSubscriptionTopics())
			return NewKafkaConsumer[T](
				kafkaCli,
				c.GetSubscriptionTopics(),
				c.GetConsumerGroupId(),
				logger,
			)
		}),
		fx.Invoke(func(lifecycle fx.Lifecycle, con Consumer[T], c T) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return con.Subscribe(c)
				},
				OnStop: func(ctx context.Context) error {
					return con.Unsubscribe(c)
				},
			})
		}),
	)
}
