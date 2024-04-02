package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type Env struct {
	Port uint16 `env:"HTTP_PORT" required:"true"`

	IotDeviceDBUri  string `env:"MONGO_URI" required:"true"`
	IotDeviceDBName string `env:"MONGO_DB_NAME" required:"true"`

	AccountCookieName string `env:"ACCOUNT_COOKIE_NAME" required:"true"`
	ClusterCookieName string `env:"CLUSTER_COOKIE_NAME" required:"true"`

	// NATS:start
	NatsURL                string `env:"NATS_URL" required:"true"`
	NatsResourceSyncStream string `env:"NATS_RESOURCE_STREAM" required:"true"`
	// NATS:end

	DefaultProjectWorkspaceName string `env:"DEFAULT_PROJECT_WORKSPACE_NAME" required:"true"`

	PromHttpAddr       string `env:"PROM_HTTP_ADDR" required:"true"`
	IsDev              bool
	KubernetesApiProxy string `env:"KUBERNETES_API_PROXY"`

	SessionKVBucket      string `env:"SESSION_KV_BUCKET" required:"true"`
	ConsoleCacheKVBucket string `env:"CONSOLE_CACHE_KV_BUCKET" required:"true"`

	DeviceNamespace string `env:"DEVICE_NAMESPACE" required:"true"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, errors.NewE(err)
	}
	return &e, nil
}
