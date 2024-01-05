package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type Env struct {
	Port                   uint16 `env:"HTTP_PORT" required:"true"`
	LogsAndMetricsHttpPort uint16 `env:"LOGS_AND_METRICS_HTTP_PORT" required:"true"`

	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	ConsoleDBUri  string `env:"MONGO_URI" required:"true"`
	ConsoleDBName string `env:"MONGO_DB_NAME" required:"true"`

	AccountCookieName string `env:"ACCOUNT_COOKIE_NAME" required:"true"`
	ClusterCookieName string `env:"CLUSTER_COOKIE_NAME" required:"true"`

	// NATS:start
	NatsURL                string `env:"NATS_URL" required:"true"`
	NatsResourceSyncStream string `env:"NATS_RESOURCE_STREAM" required:"true"`
	// NATS:end

	IAMGrpcAddr   string `env:"IAM_GRPC_ADDR" required:"true"`
	InfraGrpcAddr string `env:"INFRA_GRPC_ADDR" required:"true"`

	DefaultProjectWorkspaceName string `env:"DEFAULT_PROJECT_WORKSPACE_NAME" required:"true"`

	LokiServerHttpAddr        string `env:"LOKI_SERVER_HTTP_ADDR" required:"true"`
	PromHttpAddr              string `env:"PROM_HTTP_ADDR" required:"true"`
	SessionKVBucket           string `env:"SESSION_KV_BUCKET" required:"true"`
	ProjectClusterMapKVBucket string `env:"PROJECT_CLUSTER_MAP_KV_BUCKET" required:"true"`
	IsDev                     bool
	KubernetesApiProxy        string `env:"KUBERNETES_API_PROXY"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, errors.NewE(err)
	}
	return &e, nil
}
