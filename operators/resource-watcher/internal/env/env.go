package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD" required:"true"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES" required:"true"`

	GrpcAddr string `env:"GRPC_ADDR" required:"true"`

	AccountName string `env:"ACCOUNT_NAME" required:"true"`
	ClusterName string `env:"CLUSTER_NAME" required:"true"`
	//AccessToken string `env:"ACCESS_TOKEN" required:"true"`

	ClusterIdentitySecretName      string `env:"CLUSTER_IDENTITY_SECRET_NAME" required:"true"`
	ClusterIdentitySecretNamespace string `env:"CLUSTER_IDENTITY_SECRET_NAMESPACE" required:"true"`

	OperatorsNamespace string `env:"OPERATORS_NAMESPACE" required:"true"`
}

func GetEnvOrDie() *Env {
	ev := Env{}
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
