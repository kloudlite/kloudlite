package env

import "github.com/codingconcepts/env"

type Env struct {
	MaxConcurrentReconciles int    `env:"MAX_CONCURRENT_RECONCILES" default:"1"`
	GatewayAdminApiAddr     string `env:"GATEWAY_ADMIN_API_ADDR" required:"true"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
