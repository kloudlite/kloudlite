package app

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/fx"
)

type M map[string]interface{}

const MaxRetries = 2

func processMesssage(ctx context.Context, topic string, message []byte) error {
	return nil
}

func readMessages(kConsumer *kafka.Consumer) {
	for {
		msg, err := kConsumer.ReadMessage(-1)
		if err != nil {
			fmt.Printf("could not read kafka message because %v\n", err)
		}

		hasProcessed := false
		for rCount := 0; rCount < MaxRetries; rCount++ {
			err = processMesssage(context.TODO(), *msg.TopicPartition.Topic, msg.Value)
			if err != nil {
				fmt.Println("Retrying ....")
				continue
			}
			hasProcessed = true
		}
		if !hasProcessed {
			fmt.Printf("could not process message even after %d retries\n", MaxRetries)
		}
		kConsumer.CommitMessage(msg)
		fmt.Println("committed msg...")
	}
}

var Module = fx.Module("app",
	fx.Invoke(func(lf fx.Lifecycle, consumer *kafka.Consumer) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					readMessages(consumer)
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return consumer.Close()
			},
		})
	}),
)
