package domain

import "go.uber.org/fx"

type domain struct{}

var Module = fx.Module("domain",
	fx.Provide(func() Domain {
		return &domain{}
	}),
)
