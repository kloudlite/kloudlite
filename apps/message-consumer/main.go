package main

import (
	"flag"
	"fmt"
	"os"

	fm "kloudlite.io/apps/message-consumer/internal/framework"
)

func readEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic("env BOOTSTRAP_SERVERS not set")
	}
	return value
}

func main() {
	isDevelopment := flag.Bool("dev", false, "development mode")
	flag.Parse()

	start, err := fm.MakeFramework(&fm.Config{
		IsDev:           *isDevelopment,
		KafkaBrokers:    readEnv("BOOTSTRAP_SERVERS"),
		ConsumerGroupId: readEnv("CONSUMER_GROUP_ID"),
		TopicPrefix:     readEnv("TOPIC_PREFIX"),
	})

	if err != nil {
		panic(fmt.Errorf("failed to start framework because %v", err))
	}

	start()
}
