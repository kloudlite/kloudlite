package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int    `env:"MAX_CONCURRENT_RECONCILES" required:"true"`
	GatewayAdminHttpPort    uint16 `env:"GATEWAY_ADMIN_HTTP_PORT" required:"true"`

	GatewayAdminSvcAddr   string `env:"GATEWAY_ADMIN_SVC_ADDR" required:"true"`
	GatewayAdminNamespace string `env:"GATEWAY_ADMIN_NAMESPACE" default:"kl-gateway"`

	// ClusterCIDR            string `env:"CLUSTER_CIDR" required:"true"`
	// NumReservedForServices int    `env:"NUM_RESERVED_FOR_SERVICES" required:"true"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}

func GetEnvOrDie() *Env {
	ev, err := LoadEnv()
	if err != nil {
		panic(err)
	}
	return ev
}
