package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud/aws"
)

func main() {
	ctx, cf := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGTERM)
	defer cf()

	p, err := aws.NewProvider(ctx, os.Getenv("K3S_URL"), os.Getenv("K3S_TOKEN"))
	if err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "machine-status":
		{
			mi, err := p.GetMachineStatus(ctx, os.Args[2])
			if err != nil {
				panic(err)
			}

			fmt.Printf("%#v", *mi)
		}
	}
}
