package app

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/auth/internal/app/graph"
	"kloudlite.io/apps/auth/internal/app/graph/generated"
	"kloudlite.io/apps/auth/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/pkg/cache"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
	"net/http"
)

var Module = fx.Module("app",
	repos.NewFxMongoRepo[domain.User](domain.UserIndexes),
	repos.NewFxMongoRepo[domain.AccessToken](domain.AccessTokenIndexes),
	fx.Invoke(func(
		server *http.ServeMux,
		d domain.Domain,
		cacheClient cache.Client,
	) {
		schema := generated.NewExecutableSchema(
			generated.Config{Resolvers: graph.NewResolver(d)},
		)
		httpServer.SetupGQLServer(
			server,
			schema,
			cache.NewSessionRepo[*common.AuthSession](
				cacheClient,
				common.CookieName,
				common.CacheSessionPrefix,
			),
		)
	}),
)
