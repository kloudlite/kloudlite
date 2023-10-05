package app

import (
	"github.com/gofiber/fiber/v2"
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/infra"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/message-office-internal"

	"kloudlite.io/apps/message-office/internal/app/graph"
	"kloudlite.io/apps/message-office/internal/app/graph/generated"
	proto_rpc "kloudlite.io/apps/message-office/internal/app/proto-rpc"
	"kloudlite.io/apps/message-office/internal/domain"
	"kloudlite.io/apps/message-office/internal/env"
	"kloudlite.io/pkg/grpc"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
)

type RealVectorGrpcClient grpc.Client
type InfraGrpcClient grpc.Client

type ExternalGrpcServer grpc.Server
type InternalGrpcServer grpc.Server

var Module = fx.Module("app",
	redpanda.NewProducerFx[redpanda.Client](),

	fx.Provide(func(restCfg *rest.Config) (kubectl.ControllerClient, error) {
		scheme := runtime.NewScheme()
		if err := artifactsv1.AddToScheme(scheme); err != nil {
			return nil, err
		}
		return kubectl.NewClientWithScheme(restCfg, scheme)
	}),

	fx.Provide(func(conn InfraGrpcClient) infra.InfraClient {
		return infra.NewInfraClient(conn)
	}),

	fx.Provide(func(logger logging.Logger, producer redpanda.Producer, ev *env.Env, d domain.Domain, kControllerCli kubectl.ControllerClient, infraCli infra.InfraClient) messages.MessageDispatchServiceServer {
		return &grpcServer{
			domain:           d,
			infraClient:      infraCli,
			logger:           logger.WithKV("component", "message-dispatcher-grpc-server"),
			producer:         producer,
			consumers:        map[string]redpanda.Consumer{},
			ev:               ev,
			k8sControllerCli: kControllerCli,
		}
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

	repos.NewFxMongoRepo[*domain.MessageOfficeToken]("mo_tokens", "mot", domain.MOTokenIndexes),
	repos.NewFxMongoRepo[*domain.AccessToken]("acc_tokens", "acct", domain.AccessTokenIndexes),

	fx.Invoke(
		func(server *fiber.App, d domain.Domain) {
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
