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

	DnsHostedZone   string `env:"DNS_HOSTED_ZONE" required:"true"`
	TlsDomainPrefix string `env:"TLS_DOMAIN_PREFIX"`

	AccountName string `env:"ACCOUNT_NAME"`
	ClusterName string `env:"CLUSTER_NAME"`

	ClusterInternalDns string `env:"CLUSTER_INTERNAL_DNS" default:"cluster.local"`

	DefaultIngressClass    string `env:"DEFAULT_INGRESS_CLASS" required:"true"`
	EnvironmentIngressName string `env:"ENVIRONMENT_INGRESS_NAME" default:"env-ingress"`

	WgGatewayImage string `env:"WG_GATEWAY_IMAGE"`
	WgAgentImage   string `env:"WG_AGENT_IMAGE"`

	WgIpBase string `env:"WG_IP_BASE" default:"10.13.0.0"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
