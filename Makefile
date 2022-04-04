# consumer
consumer.run: export BOOTSTRAP_SERVERS = kafka-kafka-bootstrap.hotspot:9092

# producer
producer.run:
	go run apps/wireguard/main.go

# wireguard
wireguard.gql:
	cd apps/wireguard/internal/app && go run github.com/99designs/gqlgen generate
wireguard.run:
	go run apps/wireguard/main.go
