SHELL := bash

dev: export MONGO_URI=mongodb://localhost:27017
dev: export MONGO_DB_NAME=console-go
dev: export KAFKA_BOOTSTRAP_SERVERS=kafka-kafka-0.kafka-kafka-brokers.hotspot:9092,kafka-kafka-1.kafka-kafka-brokers.hotspot:9092
dev: export KAFKA_INFRA_TOPIC=dev-hotspot-infra
dev: export KAFKA_INFRA_ACTION_RESULT_TOPIC=dev-hotspot-infra-results
dev: export KAFKA_GROUP_ID=group-id
dev: export DO_API_KEY=***REMOVED***
dev: export DATA_PATH=/tmp/do
dev: export SSH_KEYS_PATH=/tmp/ssh
dev: _genssh
dev:
	go run ./main.go --dev

_genssh:
	# ssh-keygen -t rsa -b 4096
	# mkdir -p /tmp/do /tmp/ssh && 
	cp ~/.ssh/* /tmp/ssh
