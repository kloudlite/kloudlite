package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`

	CloudProviderName   string `env:"CLOUD_PROVIDER_NAME" required:"true"`
	CloudProviderRegion string `env:"CLOUD_PROVIDER_REGION" required:"true"`

	K3sJoinToken        string `env:"K3S_JOIN_TOKEN" required:"true"`
	K3sServerPublicHost string `env:"K3S_SERVER_PUBLIC_HOST" required:"true"`

	KloudliteAccountName string `env:"KLOUDLITE_ACCOUNT_NAME" required:"true"`
	KloudliteClusterName string `env:"KLOUDLITE_CLUSTER_NAME" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}

	return &ev
}
