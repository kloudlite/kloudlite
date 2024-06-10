package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int    `env:"MAX_CONCURRENT_RECONCILES" required:"true"`
	GatewayAdminNamespace   string `env:"GATEWAY_ADMIN_NAMESPACE" default:"kl-gateway"`
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
