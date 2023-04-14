package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/message-office/internal/app/graph"
	"kloudlite.io/apps/message-office/internal/app/graph/generated"
	"kloudlite.io/apps/message-office/internal/domain"
	"kloudlite.io/apps/message-office/internal/env"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
)

var Module = fx.Module("app",
	fx.Provide(func(logger logging.Logger, producer redpanda.Producer, ev *env.Env) messages.MessageDispatchServiceServer {
		return &grpcServer{
			logger:    logger,
			producer:  producer,
			consumers: map[string]redpanda.Consumer{},
			ev:        ev,
		}
	}),
	fx.Invoke(
		func(server *grpc.Server, messageServer messages.MessageDispatchServiceServer) {
			messages.RegisterMessageDispatchServiceServer(server, messageServer)
		},
	),
	repos.NewFxMongoRepo[*domain.MessageOfficeToken]("mo_tokens", "mot", domain.MOTokenIndexes),
	repos.NewFxMongoRepo[*domain.AccessToken]("acc_tokens", "acct", domain.AccessTokenIndexes),
	fx.Invoke(
		func(
			server *fiber.App,
			d domain.Domain,
		) {
			schema := generated.NewExecutableSchema(
				generated.Config{
					Resolvers: &graph.Resolver{Domain: d},
				},
			)
			httpServer.SetupGQLServer(server, schema)
		},
	),
	domain.Module,
)
