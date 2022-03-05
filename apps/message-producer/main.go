package main

import (
	"fmt"
	"os"
	"strconv"

	"kloudlite.io/apps/message-producer/internal/framework"
)

func main() {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(fmt.Errorf("failed to parse PORT: %v", port))
	}
	bootstrap, ok := os.LookupEnv("BOOTSTRAP_SERVERS")
	if !ok {
		panic(err)
	}
	fm, err := framework.MakeFW(&framework.Config{
		HttpPort:     port,
		KafkaBrokers: bootstrap,
	})

	if err != nil || fm == nil {
		panic(fmt.Errorf("failed to create framework: %w", err))
	}

	fm()
}
