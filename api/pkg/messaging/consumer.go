package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"kloudlite.io/pkg/errors"
)

type consumer struct {
	kafkaConsumer *kafka.Consumer
}

func (c *consumer) OnMessage(topics []string, callback func(topic string, message Json) error) (e error) {
	defer errors.HandleErr(&e)
	e = c.kafkaConsumer.SubscribeTopics(topics, nil)
	errors.AssertNoError(e, fmt.Errorf("failed to subscribe to topics %v", topics))
	go func() {
		for {
			msg, e := c.kafkaConsumer.ReadMessage(-1)
			if e != nil {
				fmt.Errorf("failed to read message from kafka")
				//continue
			}

			var message Json
			e = json.Unmarshal(msg.Value, &message)
			if e != nil {
				fmt.Errorf("unable to typecast message into json")
				//continue
			}

			e = callback(*msg.TopicPartition.Topic, message)
			if e != nil {
				e = callback(*msg.TopicPartition.Topic, message)
				if e != nil {
					fmt.Errorf("failed to process message after 2 retries")
					//continue
				}
			}
			c.kafkaConsumer.CommitMessage(msg)
		}
	}()
	return nil
}

func (c *consumer) Unsubscribe() error {
	return c.kafkaConsumer.Unsubscribe()
}

type Consumer interface {
	OnMessage(topics []string, callback func(topic string, message Json) error) error
	Unsubscribe() error
}

func NewKafkaConsumer(kafkaBrokers string, consumerGroupId string) (messenger Consumer, e error) {
	defer errors.HandleErr(&e)
	c, e := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  kafkaBrokers,
		"group.id":           consumerGroupId,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	})
	errors.AssertNoError(e, fmt.Errorf("failed to create kafka producer"))
	return &consumer{
		kafkaConsumer: c,
	}, e
}
