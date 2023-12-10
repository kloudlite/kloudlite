package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/messaging/nats"
	"kloudlite.io/pkg/messaging/types"
)

func main() {
	natsUrl := os.Getenv("NATS_URL")
	// natsStream := os.Getenv("NATS_STREAM")

	nc, err := nats.NewClient(natsUrl, nats.ClientOpts{
		Name: "nats-producer",
	})
	if err != nil {
		log.Fatal(err)
	}

	jc, err := nats.NewJetstreamClient(nc)
	if err != nil {
		log.Fatal(err)
	}

	var producer messaging.Producer = jc.CreateProducer()
	if err != nil {
		log.Fatal(err)
	}

	subjectBase := fmt.Sprintf("resource-sync.account-sasfa.cluster-asdfasf.platform.kloudlite-console.resource-update")

	fmt.Printf("subject base: %s\n", subjectBase)

	for i := 0; i < 1000; i++ {
		if err := producer.Produce(context.TODO(), types.ProduceMsg{
			Subject: func() string {
				if i%2 == 0 {
					return fmt.Sprintf("%s.even", subjectBase)
				}
				return fmt.Sprintf("%s.odd", subjectBase)
			}(),
			Payload: func() []byte {
				if i%2 == 0 {
					return []byte(fmt.Sprintf("even: %d", i))
				}
				return []byte(fmt.Sprintf("odd: %d", i))
			}(),
		}); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d message sent\n", i+1)
		time.Sleep(1 * time.Second)
	}

	producer.Stop(context.TODO())
}
