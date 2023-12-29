package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	if ev.MaxConcurrentReconciles == 0 {
		ev.MaxConcurrentReconciles = 5
	}
	return &ev
}
