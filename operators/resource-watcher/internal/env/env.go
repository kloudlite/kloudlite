package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	CommonEnv
	RunningOnPlatformEnv
	RunningOnTargetClusterEnv
}

type CommonEnv struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD" required:"true"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES" required:"true"`
	AccountName             string        `env:"ACCOUNT_NAME" required:"true"`
	ClusterName             string        `env:"CLUSTER_NAME" required:"true"`
}

type RunningOnPlatformEnv struct {
	KafkaBrokers              string `env:"KAFKA_BROKERS" required:"true"`
	KafkaResourceUpdatesTopic string `env:"KAFKA_RESOURCE_UPDATES_TOPIC" required:"true"`
	KafkaInfraUpdatesTopic    string `env:"KAFKA_INFRA_UPDATES_TOPIC" required:"true"`
}

type RunningOnTargetClusterEnv struct {
	GrpcAddr    string `env:"GRPC_ADDR" required:"true"`
	AccessToken string `env:"ACCESS_TOKEN" required:"true"`

	ClusterIdentitySecretName      string `env:"CLUSTER_IDENTITY_SECRET_NAME" required:"true"`
	ClusterIdentitySecretNamespace string `env:"CLUSTER_IDENTITY_SECRET_NAMESPACE" required:"true"`
}

func GetCommonEnv() (CommonEnv, error) {
	ev := CommonEnv{}
	if err := env.Set(&ev); err != nil {
		return CommonEnv{}, err
	}
	return ev, nil
}

func GetPlatformEnv() (RunningOnPlatformEnv, error) {
	ev := RunningOnPlatformEnv{}
	if err := env.Set(&ev); err != nil {
		return RunningOnPlatformEnv{}, err
	}
	return ev, nil
}

func GetTargetClusterEnvs() (RunningOnTargetClusterEnv, error) {
	ev := RunningOnTargetClusterEnv{}
	if err := env.Set(&ev); err != nil {
		return RunningOnTargetClusterEnv{}, err
	}
	return ev, nil
}
