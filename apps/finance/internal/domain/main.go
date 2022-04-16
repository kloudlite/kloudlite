package domain

import "go.uber.org/fx"

var Module = fx.Module("",
	fx.Provide(fxDomain),
)
