package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	GrpcAddr string `env:"GRPC_ADDR" required:"true"`

	ClusterToken               string `env:"CLUSTER_TOKEN" required:"false"`
	AccessToken                string `env:"ACCESS_TOKEN" required:"false"`
	AccessTokenSecretName      string `env:"ACCESS_TOKEN_SECRET_NAME" required:"true"`
	AccessTokenSecretNamespace string `env:"ACCESS_TOKEN_SECRET_NAMESPACE" required:"true"`

	ClusterName string `env:"CLUSTER_NAME" required:"true"`
	AccountName string `env:"ACCOUNT_NAME" required:"true"`

	ImagePullSecretName      string `env:"IMAGE_PULL_SECRET_NAME" required:"true"`
	ImagePullSecretNamespace string `env:"IMAGE_PULL_SECRET_NAMESPACE" required:"true"`

	VectorProxyGrpcServerAddr string `env:"VECTOR_PROXY_GRPC_SERVER_ADDR" required:"true"`
	ResourceWatcherName       string `env:"RESOURCE_WATCHER_NAME" required:"true"`
	ResourceWatcherNamespace  string `env:"RESOURCE_WATCHER_NAMESPACE" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
