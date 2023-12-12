package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kloudlite/api/pkg/messaging/nats"
	// "log"
	// "os"
	// "kloudlite.io/pkg/messaging"
	// "kloudlite.io/pkg/messaging/nats"
)

func main() {
	var natsUrl string
	var stream string

	flag.StringVar(&natsUrl, "url", "", "--url")
	flag.StringVar(&stream, "stream", "", "--stream")
	flag.Parse()

	nc, err := nats.NewClient(natsUrl, nats.ClientOpts{Name: "nats-manager-cli"})
	if err != nil {
		panic(err)
	}

	jc, err := nats.NewJetstreamClient(nc)
	if err != nil {
		panic(err)
	}

	args := flag.CommandLine.Args()

	if len(args) == 1 {
		args = append(args, "")
	}

	switch args[0] {
	case "consumers":
		{
			switch args[1] {
			case "ls":
				{
					ci, err := jc.ListConsumers(context.TODO(), stream)
					if err != nil {
						log.Fatal(err)
					}
					for i := range ci {
						fmt.Printf("Name: %s\tSubject: %+v\n", ci[i].Name, ci[i].Config.FilterSubjects)
					}
				}

			case "rm":
				{
					ci, err := jc.ListConsumers(context.TODO(), stream)
					if err != nil {
						log.Fatal(err)
					}
					for i := range ci {
						fmt.Printf("Name: %s\tSubject: %+v\n", ci[i].Name, ci[i].Config.FilterSubjects)
					}
					fmt.Printf("\nEnter consumer name: ")
					var consumerName string
					fmt.Scanf("%s", &consumerName)
					err = jc.DeleteConsumer(context.TODO(), stream, consumerName)
					if err != nil {
						log.Fatal(err)
					}
				}
			default:
				{
					fmt.Printf("incorrect usage, valid subcommands: [ls]\n")
					os.Exit(1)
				}
			}
		}
	default:
		{
			fmt.Printf("incorrect usage, valid commands: [consumers]\n")
			os.Exit(1)
		}
	}

	fmt.Printf("args: %+v\n", args)

	// nc, err := nats.NewClient(natsUrl, nats.ClientOpts{
	// 	Name: "nats-manager-cli",
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// jc, err := nats.NewJetstreamClient(nc)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// var producer messaging.Producer = jc.CreateProducer()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// subjectBase := fmt.Sprintf("resource-sync.account-sasfa.cluster-asdfasf.platform.kloudlite-console.resource-update")
	//
	// fmt.Printf("subject base: %s\n", subjectBase)
}
