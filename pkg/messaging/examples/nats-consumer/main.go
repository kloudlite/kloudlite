package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/messaging/nats"
	"kloudlite.io/pkg/messaging/types"
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

	// subjectBase := fmt.Sprintf("%s.account-*.cluster-*.platform.kloudlite-console.resource-update", natsStream)
	subjectBase := fmt.Sprintf("resource-sync.*.*.platform.kloudlite-console.resource-update")
	_ = subjectBase

	consumer, err = jc.CreateConsumer(context.TODO(), nats.JetstreamConsumerArgs{
		Stream: natsStream,
		ConsumerConfig: nats.ConsumerConfig{
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
	})
}
