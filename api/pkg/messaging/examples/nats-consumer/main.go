package main

import (
	"context"
	"log"

	"github.com/nats-io/nats.go/jetstream"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/messaging/nats"
	"kloudlite.io/pkg/messaging/types"
)

func main() {
	var consumer messaging.Consumer

	nc, err := nats.NewClient("tls://connect.ngs.global", nats.ClientOpts{
		CrdeentialsFile: "/home/nxtcoder17/Downloads/NGS-Default-CLI.creds",
		Options: nats.Options{
			Servers: []string{"tls://connect.ngs.global"},
			Name:    "nats-producer",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	jc, err := nats.NewJetstreamClient(nc)
	if err != nil {
		log.Fatal(err)
	}

	consumer, err = nats.NewJetstreamConsumer(context.TODO(), jc, nats.JetstreamConsumerArgs{
		Stream: "example",
		ConsumerConfig: jetstream.ConsumerConfig{
		  // FilterSubjects: []string{"example.even"},
		  FilterSubject: "example.even",
			Name:        "example-consumer",
			Durable:     "example-consumer",
			Description: "this is a test consumer",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	consumer.Consume(context.TODO(), func(msg *types.ConsumeMsg) error {
		log.Println(string(msg.NatsJetstreamMsg.Payload))
		return nil
	})
}
