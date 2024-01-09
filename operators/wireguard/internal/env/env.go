package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`

	ClusterPodCidr     string `env:"CLUSTER_POD_CIDR" required:"true"` // default 10.42.0.0/16
	ClusterServiceCidr string `env:"CLUSTER_SVC_CIDR" required:"true"` // default 10.43.0.0/16

	DnsHostedZone string `env:"DNS_HOSTED_ZONE" required:"true"`

	ClusterInternalDns  string `env:"CLUSTER_INTERNAL_DNS" default:"cluster.local"`
	DeviceInfoNamespace string `env:"DEVICE_INFO_NAMESPACE" default:"device-info"`

	EnvironmentIngressName string `env:"ENVIRONMENT_INGRESS_NAME" default:"env-ingress"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
