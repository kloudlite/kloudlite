package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logger"
)

type Message []byte

func (m Message) Unmarshal(x any) error {
	return json.Unmarshal(m, &x)
}

type ConsumerCallback func(context context.Context, topic string, message Message) error

type consumer struct {
	kafkaConsumer *kafka.Consumer
	topics        []string
	callback      ConsumerCallback
	stopChan      chan bool
	logger        logger.Logger
}

func (c *consumer) Unsubscribe(context.Context) error {
	c.stopChan <- true
	return c.kafkaConsumer.Unsubscribe()
}

func (c *consumer) Subscribe(cc context.Context) error {
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
			fmt.Println("here")
			if stop {
				return
			}
			msg, e := c.kafkaConsumer.ReadMessage(-1)
			if e != nil {
				c.logger.Errorf("could not read kafka message because %v", e)
				continue
			}
			fmt.Println("Received message: ", string(msg.Value))
			e = c.callback(context.TODO(), *msg.TopicPartition.Topic, msg.Value)

			if e != nil {
				e = c.callback(context.TODO(), *msg.TopicPartition.Topic, msg.Value)
				if e != nil {
					fmt.Errorf("failed to process message after 2 retries")
				}
			}
			c.kafkaConsumer.CommitMessage(msg)
		}
	}()
	return nil
}

type Consumer interface {
	Unsubscribe(context context.Context) error
	Subscribe(context context.Context) error
}

func NewKafkaConsumer(
	kafkaCli KafkaClient,
	topics []string,
	consumerGroupId string,
	logger logger.Logger,
	callback ConsumerCallback,
) (messenger Consumer, e error) {
	defer errors.HandleErr(&e)
	c, e := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  kafkaCli.GetBrokers(),
		"group.id":           consumerGroupId,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	})
	errors.AssertNoError(e, fmt.Errorf("failed to create kafka producer"))
	return &consumer{
		kafkaConsumer: c,
		topics:        topics,
		callback:      callback,
		logger:        logger,
	}, e
}
