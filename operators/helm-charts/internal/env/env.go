package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	if ev.ReconcilePeriod == 0 {
		ev.ReconcilePeriod = 10 * time.Second
	}
	if ev.MaxConcurrentReconciles == 0 {
		ev.MaxConcurrentReconciles = 5
	}
	return &ev
}
