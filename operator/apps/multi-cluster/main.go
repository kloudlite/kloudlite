package main

import (
	"fmt"

	"github.com/kloudlite/operator/apps/multi-cluster/apps/client"
	"github.com/kloudlite/operator/apps/multi-cluster/apps/server"
	"github.com/kloudlite/operator/apps/multi-cluster/flags"
)

func main() {
	if err := Run(); err != nil {
		panic(err)
	}
}

func Run() error {
	fmt.Println("Running", flags.Mode)
	if flags.Mode == "server" {
		return server.Run()
	}

	return client.Run()
}
