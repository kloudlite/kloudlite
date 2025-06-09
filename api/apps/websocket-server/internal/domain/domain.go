package domain

import (
	"context"

	"github.com/gofiber/websocket/v2"

	"github.com/kloudlite/api/apps/websocket-server/internal/env"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
	"go.uber.org/fx"
)

type SocketService interface {
	HandleWebSocket(ctx context.Context, c *websocket.Conn) error
}

type Domain interface {
	SocketService
}

type domain struct {
	iamClient       iam.IAMClient
	natsClient      *nats.Client
	jetStreamClient *nats.JetstreamClient
	env             *env.Env

	logger logging.Logger
}

func NewDomain(
	iamCli iam.IAMClient,
	env *env.Env,

	logger logging.Logger,
	natsClient *nats.Client,
	jetStreamClient *nats.JetstreamClient,
) Domain {
	return &domain{
		iamClient:       iamCli,
		natsClient:      natsClient,
		jetStreamClient: jetStreamClient,

		env:    env,
		logger: logger,
	}
}

var Module = fx.Module("domain", fx.Provide(NewDomain))
