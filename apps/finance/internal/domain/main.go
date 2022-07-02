package domain

import (
	"go.uber.org/fx"
	"kloudlite.io/pkg/config"
)

type Env struct {
	InventoryPath string `env:"INVENTORY_PATH"`
}

var Module = fx.Module("domain",
	fx.Provide(fxDomain),
	config.EnvFx[Env](),
)
