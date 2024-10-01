package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type Env struct {
	Port     uint16 `env:"HTTP_PORT" required:"true"`
	GrpcPort uint16 `env:"GRPC_PORT" required:"true"`

	DNSAddr            string `env:"DNS_ADDR" required:"true"`
	KloudliteDNSSuffix string `env:"KLOUDLITE_DNS_SUFFIX" required:"true"`
	ConsoleDBUri       string `env:"MONGO_URI" required:"true"`
	ConsoleDBName      string `env:"MONGO_DB_NAME" required:"true"`

	AccountCookieName string `env:"ACCOUNT_COOKIE_NAME" required:"true"`
	ClusterCookieName string `env:"CLUSTER_COOKIE_NAME" required:"true"`

	NatsURL                    string `env:"NATS_URL" required:"true"`
	NatsReceiveFromAgentStream string `env:"NATS_RECEIVE_FROM_AGENT_STREAM" required:"true"`
	EventsNatsStream           string `env:"EVENTS_NATS_STREAM" required:"true"`
	WebhookTokenHashingSecret  string `env:"WEBHOOK_TOKEN_HASHING_SECRET" required:"true"`
	WebhookURL                 string `env:"WEBHOOK_URL" required:"true"`

	IAMGrpcAddr                   string `env:"IAM_GRPC_ADDR" required:"true"`
	InfraGrpcAddr                 string `env:"INFRA_GRPC_ADDR" required:"true"`
	MessageOfficeInternalGRPCAddr string `env:"MESSAGE_OFFICE_INTERNAL_GRPC_ADDR" required:"true"`
	AccountGRPCAddr               string `env:"ACCOUNT_GRPC_ADDR" required:"true"`

	SessionKVBucket      string `env:"SESSION_KV_BUCKET" required:"true"`
	ConsoleCacheKVBucket string `env:"CONSOLE_CACHE_KV_BUCKET" required:"true"`
	IsDev                bool

	KubernetesApiProxy string `env:"KUBERNETES_API_PROXY" default:"localhost:8080"`

	DefaultEnvTemplateAccountName string `env:"DEFAULT_ENV_TEMPLATE_ACCOUNT_NAME"`
	DefaultEnvTemplateName        string `env:"DEFAULT_ENV_TEMPLATE_NAME"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, errors.NewE(err)
	}
	return &e, nil
}
