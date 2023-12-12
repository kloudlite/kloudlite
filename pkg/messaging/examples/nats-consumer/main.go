package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/kloudlite/api/pkg/messaging"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/api/pkg/nats"
)

func main() {
	var consumer messaging.Consumer

	natsUrl := os.Getenv("NATS_URL")
	natsStream := os.Getenv("NATS_STREAM")

	nc, err := nats.NewClient(natsUrl, nats.ClientOpts{
		Name: "nats-consumer",
	})
	if err != nil {
		log.Fatal(err)
	}

	jc, err := nats.NewJetstreamClient(nc)
	if err != nil {
		log.Fatal(err)
	}

	subjectBase := fmt.Sprintf("resource-sync.*.*.platform.kloudlite-console.resource-update")
	_ = subjectBase

	consumer, err = msg_nats.NewJetstreamConsumer(context.TODO(), jc, msg_nats.JetstreamConsumerArgs{
		Stream: natsStream,
		ConsumerConfig: msg_nats.ConsumerConfig{
			FilterSubjects: []string{subjectBase},
			Name:           "example-consumer",
			Durable:        "example-consumer",
			Description:    "this is a test consumer",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	consumer.Consume(func(msg *types.ConsumeMsg) error {
		log.Println(string(msg.Payload))
		return nil
	}, types.ConsumeOpts{})
}
