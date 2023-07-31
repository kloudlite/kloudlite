package app

import (
	"github.com/gofiber/fiber/v2"
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	"github.com/kloudlite/operator/grpc-interfaces/grpc/messages"
	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"kloudlite.io/apps/message-office/internal/app/graph"
	"kloudlite.io/apps/message-office/internal/app/graph/generated"
	proto_rpc "kloudlite.io/apps/message-office/internal/app/proto-rpc"
	"kloudlite.io/apps/message-office/internal/domain"
	"kloudlite.io/apps/message-office/internal/env"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
)

type ContainerRegistryGrpcConnection *grpc.ClientConn
type RealVectorGrpcClientConn *grpc.ClientConn

var Module = fx.Module("app",
	redpanda.NewProducerFx[redpanda.Client](),

	fx.Provide(func(restCfg *rest.Config) (kubectl.ControllerClient, error) {
		scheme := runtime.NewScheme()
		if err := artifactsv1.AddToScheme(scheme); err != nil {
			return nil, err
		}
		return kubectl.NewClientWithScheme(restCfg, scheme)
	}),

	fx.Provide(func(logger logging.Logger, producer redpanda.Producer, ev *env.Env, d domain.Domain, kControllerCli kubectl.ControllerClient) messages.MessageDispatchServiceServer {
		return &grpcServer{
			domain:           d,
			logger:           logger.WithKV("component", "message-dispatcher-grpc-server"),
			producer:         producer,
			consumers:        map[string]redpanda.Consumer{},
			ev:               ev,
			k8sControllerCli: kControllerCli,
		}
	}),

	fx.Provide(func(conn RealVectorGrpcClientConn) proto_rpc.VectorClient {
		return proto_rpc.NewVectorClient((*grpc.ClientConn)(conn))
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

	fx.Invoke(
		func(server *grpc.Server, messageServer messages.MessageDispatchServiceServer) {
			messages.RegisterMessageDispatchServiceServer(server, messageServer)
		},
	),

	fx.Invoke(
		func(server *grpc.Server, vectorServer proto_rpc.VectorServer) {
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
