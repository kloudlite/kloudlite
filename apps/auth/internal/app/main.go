package app

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"kloudlite.io/apps/auth/internal/app/graph"
	"kloudlite.io/apps/auth/internal/app/graph/generated"
	"kloudlite.io/apps/auth/internal/domain"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
	"net/http"
)

var Module = fx.Module("app",
	fx.Provide(func(db *mongo.Database) repos.DbRepo[domain.Token] {
		return repos.NewMongoRepoAdapter[domain.Token](db, "users", "usr")
	}),
	fx.Provide(func(db *mongo.Database) repos.DbRepo[domain.User] {

		return repos.NewMongoRepoAdapter[domain.User](db, "tokens", "tkn")
	}),
	fx.Invoke(func(mux *http.ServeMux, d domain.Domain) {
		schema := generated.NewExecutableSchema(generated.Config{
			Resolvers: graph.NewResolver(d),
		})
		httpServer.SetupGQLServer(mux, schema)
	}),
)
