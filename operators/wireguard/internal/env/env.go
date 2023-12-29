package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`

	// default 10.42.0.0/16
	WGPodCidr string `env:"WG_POD_CIDR"`
	// default 10.43.0.0/16
	WGServiceCidr string `env:"WG_SVC_CIDR"`
	// default dns.khost.dev
	DnsHostedZone string `env:"DNS_HOSTED_ZONE"`

	ClusterInternalDns  string `env:"CLUSTER_INTERNAL_DNS"`
	DeviceInfoNamespace string `env:"DEVICE_INFO_NAMESPACE" default:"device-info"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
