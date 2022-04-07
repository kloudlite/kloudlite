#infra
infra.run: export SSH_KEYS_PATH=/tmp/kloudlite/keys
infra.run: export KAFKA_INFRA_ACTION_RESULT_TOPIC=kloudlite-infra-action-result
infra.run: export DO_API_KEY=***REMOVED***
infra.run: export DATA_PATH=/tmp/kloudlite/data
infra.run:
	cd apps/infra && go run main.go

# consumer
consumer.run: export BOOTSTRAP_SERVERS = kafka-kafka-bootstrap.hotspot:9092
consumer.run:
	echo "running consumer"

# producer
producer.run:
	cd apps/message-producer && go run main.go

wireguard.gql:
	cd apps/wireguard/internal/app && go run github.com/99designs/gqlgen generate
wireguard.run:
	cd apps/wireguard && go run main.go

console:
	make -C apps/console ${@:1}

infra:
	make -C apps/infra ${@:1}
