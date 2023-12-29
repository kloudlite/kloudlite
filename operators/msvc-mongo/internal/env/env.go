package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	IsDev                   bool
	MaxConcurrentReconciles int    `env:"MAX_CONCURRENT_RECONCILES"`
	ClusterInternalDNS      string `env:"CLUSTER_INTERNAL_DNS"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	if ev.ClusterInternalDNS == "" {
		ev.ClusterInternalDNS = "cluster.local"
	}
	return &ev
}
