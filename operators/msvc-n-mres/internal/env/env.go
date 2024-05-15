package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`

	MsvcCredsSvcHttpPort    uint16 `env:"MSVC_CREDS_SVC_HTTP_PORT"`
	MsvcCredsSvcRequestPath string `env:"MSVC_CREDS_SVC_REQUEST_PATH"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
