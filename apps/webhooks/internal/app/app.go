package app

import (
	"github.com/kloudlite/api/pkg/redpanda"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"app",
	redpanda.NewProducerFx[redpanda.Client](),

	LoadGitWebhook(),
	PublisherFX(),
)
