package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/messaging/nats"
	"kloudlite.io/pkg/messaging/types"
)

func main() {
	var producer messaging.Producer

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

	producer, err = nats.NewJetstreamProducer(jc)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		producer.Produce(context.TODO(), types.ProduceMsg{
			NatsJetstreamMsg: &types.NatsJetstreamProduceMsg{
				Subject: func() string {
				  if i % 2 == 0 {
				    return "example.even"
				  }
				  return "example.odd"
				}(),
				Payload: func() []byte {
				  if i % 2 == 0 {
            return []byte(fmt.Sprintf("even: %d", i))
          }
          return []byte(fmt.Sprintf("odd: %d", i))
				}(),
			},
		})
		fmt.Printf("%d message sent\n", i+1)
		time.Sleep(1 * time.Second)
	}

	producer.LifecycleOnStop(context.TODO())
}
