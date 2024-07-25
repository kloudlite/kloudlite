package app

import (
	"context"
	"log/slog"

	"github.com/kloudlite/api/apps/message-office/internal/app/graph"
	"github.com/kloudlite/api/apps/message-office/internal/app/graph/generated"
	proto_rpc "github.com/kloudlite/api/apps/message-office/internal/app/proto-rpc"
	"github.com/kloudlite/api/apps/message-office/internal/domain"
	"github.com/kloudlite/api/apps/message-office/internal/env"
	"github.com/kloudlite/api/grpc-interfaces/infra"
	message_office_internal "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
	"github.com/kloudlite/api/pkg/grpc"
	httpServer "github.com/kloudlite/api/pkg/http-server"
	"github.com/kloudlite/api/pkg/logging"
	msg_nats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"go.uber.org/fx"
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

	fx.Provide(
		func(conn InfraGRPCClient) infra.InfraClient {
			return infra.NewInfraClient(conn)
		},
	),

	fx.Provide(func(logger *slog.Logger, jc *nats.JetstreamClient, producer UpdatesProducer, ev *env.Env, d domain.Domain, infraCli infra.InfraClient) (messages.MessageDispatchServiceServer, error) {
		return NewMessageOfficeServer(producer, jc, ev, d, logger.With("component", "message-office"), infraCli)
	}),

	fx.Provide(func(conn RealVectorGrpcClient) proto_rpc.VectorClient {
		return proto_rpc.NewVectorClient(conn)
	}),

	fx.Provide(func(vectorGrpcClient proto_rpc.VectorClient, logger *slog.Logger, d domain.Domain, ev *env.Env) proto_rpc.VectorServer {
		return &vectorProxyServer{
			realVectorClient:   vectorGrpcClient,
			logger:             logger,
			domain:             d,
			tokenHashingSecret: ev.TokenHashingSecret,
			pushEventsCounter:  0,
		}
	}),

	fx.Provide(newInternalMsgServer),

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
			schema := generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{Domain: d}})
			server.SetupGraphqlServer(schema)
		},
	),
)
