package framework

import (
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"kloudlite.io/apps/message-producer/internal/app"
	"kloudlite.io/pkg/errors"
)

func makeKafkaMessagingClient(kafkaBrokers string) (messenger *app.Messenger, e error) {
	defer errors.HandleErr(&e)

	producer, e := kafka.NewProducer(
		&kafka.ConfigMap{
			"bootstrap.servers": kafkaBrokers,
		},
	)
	errors.AssertNoError(e, fmt.Errorf("failed to create kafka producer"))

	messenger = &app.Messenger{
		SendMessage: func(topic string, key string, message app.Json) (e error) {
			msgBody, e := json.Marshal(message)
			errors.AssertNoError(e, fmt.Errorf("failed to marshal message"))

			return producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: kafka.PartitionAny,
				},
				Key:   []byte(key),
				Value: msgBody,
			}, nil)
		},
	}

	return
}
