package messaging

import (
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logger"
)

type consumer[T any] struct {
	kafkaConsumer *kafka.Consumer
	topics        []string
	callback      func(topic string, message T) error
	stopChan      chan bool
	logger        logger.Logger
}

func (c *consumer[T]) Unsubscribe() error {
	c.stopChan <- true
	return c.kafkaConsumer.Unsubscribe()
}

func (c *consumer[T]) Subscribe() error {
	c.stopChan = make(chan bool, 1)
	e := c.kafkaConsumer.SubscribeTopics(c.topics, nil)
	if e != nil {
		return errors.New(fmt.Sprintf("could not subscribe to given topics %v", c.topics))
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
			}

			var message T
			fmt.Printf("Msg: %v %T, msg.Value %v %T\n", message, message, msg, msg)
			e = json.Unmarshal(msg.Value, &message)
			if e != nil {
				c.logger.Errorf("could not read kafka message because %v", e)
				//continue
			}

			e = c.callback(*msg.TopicPartition.Topic, message)
			if e != nil {
				e = c.callback(*msg.TopicPartition.Topic, message)
				if e != nil {
					fmt.Errorf("failed to process message after 2 retries")
					//continue
				}
			}
			fmt.Println("committed msg...")
			c.kafkaConsumer.CommitMessage(msg)
		}
	}()
	return nil
}

type Consumer interface {
	Unsubscribe() error
	Subscribe() error
}

func NewKafkaConsumer[T any](
	kafkaCli KafkaClient,
	topics []string,
	consumerGroupId string,
	logger logger.Logger,
	callback func(topic string, msg T) error,
) (messenger Consumer, e error) {
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
		callback:      callback,
		logger:        logger,
	}, e
}
