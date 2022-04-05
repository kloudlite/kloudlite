#infra
infra.run:
	cd apps/infra && go run main.go

# consumer
consumer.run: export BOOTSTRAP_SERVERS = kafka-kafka-bootstrap.hotspot:9092

# producer
producer.run:
	cd apps/message-producer && go run main.go

# wireguard
wireguard.gql:
	cd apps/wireguard/internal/app && go run github.com/99designs/gqlgen generate
wireguard.run:
	cd apps/wireguard && go run main.go
