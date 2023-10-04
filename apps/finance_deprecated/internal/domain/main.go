package domain

import (
	"go.uber.org/fx"
)

//	type Env struct {
//	  InventoryPath                 string `env:"INVENTORY_PATH"`
//	  CurrClusterConfigNS           string `env:"CURR_CLUSTER_CONFIG_NAMESPACE" required:"true"`
//	  CurrClusterConfigName         string `env:"CURR_CLUSTER_CONFIG_NAME" required:"true"`
//	  CurrClusterConfigClusterIdKey string `env:"CURR_CLUSTER_CONFIG_CLUSTER_ID_KEY" required:"true"`
//	}
var Module = fx.Module("domain",
	fx.Provide(fxDomain),
	// config.EnvFx[Env](),
)
