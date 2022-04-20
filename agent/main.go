package agent

import (
	"go.uber.org/fx"
	"operators.kloudlite.io/agent/internal/app"
)

func App() fx.Option {
	return app.Module
}
