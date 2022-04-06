SHELL := bash

gql: 
	cd internal/app && go run github.com/99designs/gqlgen generate

dev: export MONGO_URI=mongodb://localhost:27017
dev: export MONGO_DB_NAME=console-go
dev: export PORT=4001
dev: export DEV=true
dev: export KAFKA_BOOTSTRAP_SERVERS=kafka-kafka-0.kafka-kafka-brokers.hotspot:9092,kafka-kafka-1.kafka-kafka-brokers.hotspot:9092
dev: export KAFKA_INFRA_TOPIC=dev-hotspot-infra
dev: export ORIGINS=http://localhost:4001,https://studio.apollographql.com
dev:
	go run ./main.go --dev
