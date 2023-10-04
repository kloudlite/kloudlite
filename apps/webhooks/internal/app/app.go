package app

import (
	"go.uber.org/fx"
	"kloudlite.io/pkg/redpanda"
)

var Module = fx.Module(
	"app",
	redpanda.NewProducerFx[redpanda.Client](),

	LoadGitWebhook(),
	PublisherFX(),
)
