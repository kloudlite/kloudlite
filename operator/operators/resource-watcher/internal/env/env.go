package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	IsDev                   bool   `env:"IS_DEV"`
	MaxConcurrentReconciles int    `env:"MAX_CONCURRENT_RECONCILES" required:"true"`
	AccountName             string `env:"ACCOUNT_NAME" required:"true"`
	ClusterName             string `env:"CLUSTER_NAME" required:"true"`
	DeviceInfoNamespace     string `env:"DEVICE_INFO_NAMESPACE"`
	GrpcAddr                string `env:"GRPC_ADDR" required:"true"`
	GrpcSecureConnect       bool   `env:"GRPC_SECURE_CONNECT" default:"true"`

	AccessToken                    string `env:"ACCESS_TOKEN" required:"true"`
	ClusterIdentitySecretName      string `env:"CLUSTER_IDENTITY_SECRET_NAME" required:"true"`
	ClusterIdentitySecretNamespace string `env:"CLUSTER_IDENTITY_SECRET_NAMESPACE" required:"true"`
}

func GetEnv() (*Env, error) {
	ev := Env{}
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
