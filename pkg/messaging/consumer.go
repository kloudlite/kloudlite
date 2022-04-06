package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"kloudlite.io/pkg/errors"
)

type consumer[T any] struct {
	kafkaConsumer *kafka.Consumer
	topics        []string
	callback      func(topic string, message T) error
	stopChan      chan bool
}

// func (c *consumer[T]) OnMessage(callback func(topic string, message T) error) (e error) {
// 	defer errors.HandleErr(&e)
// 	c.callback = callback
// 	c.stopChan = make(chan int)
// 	return nil
// }

func (c *consumer[T]) Unsubscribe() error {
	c.stopChan <- true
	return c.kafkaConsumer.Unsubscribe()
}

func (c *consumer[T]) Subscribe() error {
	c.stopChan = make(chan bool, 1)
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
				fmt.Errorf("failed to read message from kafka")
				//continue
			}

			var message T
			e = json.Unmarshal(msg.Value, &message)
			if e != nil {
				fmt.Errorf("unable to typecast message into json")
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
			c.kafkaConsumer.CommitMessage(msg)
		}
	}()

	return c.kafkaConsumer.SubscribeTopics(c.topics, nil)
}

type Consumer[T any] interface {
	// OnMessage(callback func(topic string, message T) error) error
	Unsubscribe() error
	Subscribe() error
}

func NewKafkaConsumer[T any](topics []string, kafkaBrokers string, consumerGroupId string, callback func(topic string, msg T) error) (messenger Consumer[T], e error) {
	defer errors.HandleErr(&e)
	c, e := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  kafkaBrokers,
		"group.id":           consumerGroupId,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	})
	errors.AssertNoError(e, fmt.Errorf("failed to create kafka producer"))
	return &consumer[T]{
		kafkaConsumer: c,
		topics:        topics,
		callback:      callback,
	}, e
}
