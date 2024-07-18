package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	IsDev                   bool
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`

	ClusterInternalDNS string `env:"CLUSTER_INTERNAL_DNS" default:"cluster.local"`
	GlobalVpnDNS       string `env:"GLOBAL_VPN_DNS"`

	KloudliteDNSSuffix string `env:"KLOUDLITE_DNS_SUFFIX" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
