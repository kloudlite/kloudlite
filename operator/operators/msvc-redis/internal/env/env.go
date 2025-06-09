package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES" required:"true"`

	ClusterInternalDNS string `env:"CLUSTER_INTERNAL_DNS" default:"cluster.local"`
	GlobalVpnDNS       string `env:"GLOBAL_VPN_DNS"`

	KloudliteDNSSuffix string `env:"KLOUDLITE_DNS_SUFFIX" required:"true"`
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
