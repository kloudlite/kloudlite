package app

import (
	"github.com/kloudlite/api/pkg/messaging"
	msgnats "github.com/kloudlite/api/pkg/messaging/nats"
	"github.com/kloudlite/api/pkg/nats"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"app",
	fx.Provide(func(client *nats.JetstreamClient) messaging.Producer {
		return msgnats.NewJetstreamProducer(client)
	}),

	LoadGitWebhook(),
)
