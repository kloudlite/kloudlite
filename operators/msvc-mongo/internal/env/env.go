package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	IsDev                   bool
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`

	ClusterInternalDNS string `env:"CLUSTER_INTERNAL_DNS"`
	GlobalVpnDNS       string `env:"GLOBAL_VPN_DNS"`

	// MsvcCredsSvcName        string `env:"MSVC_CREDS_SVC_NAME" required:"true"`
	// MsvcCredsSvcNamespace   string `env:"MSVC_CREDS_SVC_NAMESPACE" required:"true"`
	// MsvcCredsSvcRequestPath string `env:"MSVC_CREDS_SVC_REQUEST_PATH" required:"true"`
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
