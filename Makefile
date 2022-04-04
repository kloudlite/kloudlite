wireguard.gql:
	cd apps/wireguard/internal/app && go run github.com/99designs/gqlgen

wireguard.run: export MONGO_URI = mongodb://localhost:27017
wireguard.run: export MONGO_DB_NAME = sample
wireguard.run: export PORT = 3000
wireguard.run: export IS_DEV = true
wireguard.run:
	go run apps/wireguard/main.go
