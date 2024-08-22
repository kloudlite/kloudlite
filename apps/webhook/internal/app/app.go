package app

import (
	"github.com/kloudlite/api/apps/webhook/internal/domain"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/pkg/grpc"
	"github.com/kloudlite/api/pkg/messaging"
	msgnats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/nats"
	"go.uber.org/fx"
)

type CommsGrpcClient grpc.Client

var Module = fx.Module(
	"app",
	fx.Provide(func(client *nats.JetstreamClient) messaging.Producer {
		return msgnats.NewJetstreamProducer(client)
	}),

	fx.Provide(
		func(conn CommsGrpcClient) comms.CommsClient {
			return comms.NewCommsClient(conn)
		},
	),

	domain.Module,

	LoadGitWebhook(),
)
