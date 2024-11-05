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
  return &ev
}
