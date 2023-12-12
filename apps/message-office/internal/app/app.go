package app

import (
	"context"

	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"go.uber.org/fx"
	message_office_internal "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
	msg_nats "kloudlite.io/pkg/messaging/nats"
	"kloudlite.io/pkg/nats"
	"kloudlite.io/pkg/repos"

	"kloudlite.io/apps/message-office/internal/app/graph"
	"kloudlite.io/apps/message-office/internal/app/graph/generated"
	proto_rpc "kloudlite.io/apps/message-office/internal/app/proto-rpc"
	"kloudlite.io/apps/message-office/internal/domain"
	"kloudlite.io/apps/message-office/internal/env"
	"kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
)

type (
	RealVectorGrpcClient grpc.Client
)

type (
	ExternalGrpcServer grpc.Server
	InternalGrpcServer grpc.Server
)

var Module = fx.Module("app",
	repos.NewFxMongoRepo[*domain.MessageOfficeToken]("mo_tokens", "mot", domain.MOTokenIndexes),
	repos.NewFxMongoRepo[*domain.AccessToken]("acc_tokens", "acct", domain.AccessTokenIndexes),

	fx.Provide(func(jsc *nats.JetstreamClient, logger logging.Logger) UpdatesProducer {
		return msg_nats.NewJetstreamProducer(jsc)
	}),

	fx.Invoke(func(lf fx.Lifecycle, producer UpdatesProducer) {
		lf.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return producer.Stop(ctx)
			},
		})
	}),

	domain.Module,

	fx.Provide(func(logger logging.Logger, jc *nats.JetstreamClient, producer UpdatesProducer, ev *env.Env, d domain.Domain) (messages.MessageDispatchServiceServer, error) {
		return NewMessageOfficeServer(producer, jc, ev, d, logger.WithName("message-office-grpc-server"))
	}),

	fx.Provide(func(conn RealVectorGrpcClient) proto_rpc.VectorClient {
		return proto_rpc.NewVectorClient(conn)
	}),

	fx.Provide(func(vectorGrpcClient proto_rpc.VectorClient, logger logging.Logger, d domain.Domain, ev *env.Env) proto_rpc.VectorServer {
		return &vectorProxyServer{
			realVectorClient:   vectorGrpcClient,
			logger:             logger.WithKV("component", "vector-proxy-grpc-server"),
			domain:             d,
			tokenHashingSecret: ev.TokenHashingSecret,
			pushEventsCounter:  0,
			healthCheckCounter: 0,
		}
	}),

	fx.Provide(func(d domain.Domain) message_office_internal.MessageOfficeInternalServer {
		return newInternalMsgServer(d)
	}),

	fx.Invoke(func(server InternalGrpcServer, internalMsgServer message_office_internal.MessageOfficeInternalServer) {
		message_office_internal.RegisterMessageOfficeInternalServer(server, internalMsgServer)
	}),

	fx.Invoke(
		func(server ExternalGrpcServer, messageServer messages.MessageDispatchServiceServer) {
			messages.RegisterMessageDispatchServiceServer(server, messageServer)
		},
	),

	fx.Invoke(
		func(server ExternalGrpcServer, vectorServer proto_rpc.VectorServer) {
			proto_rpc.RegisterVectorServer(server, vectorServer)
		},
	),

	fx.Invoke(
		func(server httpServer.Server, d domain.Domain) {
			schema := generated.NewExecutableSchema(
				generated.Config{
					Resolvers: &graph.Resolver{Domain: d},
				},
			)
			server.SetupGraphqlServer(schema)
		},
	),
)
