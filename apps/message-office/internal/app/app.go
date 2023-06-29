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
	"kloudlite.io/apps/message-office/internal/domain"
	"kloudlite.io/apps/message-office/internal/env"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
)

type ContainerRegistryGrpcConnection *grpc.ClientConn

var Module = fx.Module("app",
	redpanda.NewProducerFx[redpanda.Client](),

	fx.Provide(func(restCfg *rest.Config) (kubectl.ControllerClient, error) {
		scheme := runtime.NewScheme()
		artifactsv1.AddToScheme(scheme)
		return kubectl.NewClientWithScheme(restCfg, scheme)
	}),

	fx.Provide(func(logger logging.Logger, producer redpanda.Producer, ev *env.Env, d domain.Domain, kControllerCli kubectl.ControllerClient) messages.MessageDispatchServiceServer {
		return &grpcServer{
			domain:           d,
			logger:           logger,
			producer:         producer,
			consumers:        map[string]redpanda.Consumer{},
			ev:               ev,
			k8sControllerCli: kControllerCli,
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
