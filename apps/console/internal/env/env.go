package env

import "github.com/codingconcepts/env"

type Env struct {
	Port                   uint16 `env:"HTTP_PORT" required:"true"`
	LogsAndMetricsHttpPort uint16 `env:"LOGS_AND_METRICS_HTTP_PORT" required:"true"`

	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	ConsoleDBUri  string `env:"CONSOLE_DB_URI" required:"true"`
	ConsoleDBName string `env:"CONSOLE_DB_NAME" required:"true"`

	AuthRedisHosts    string `env:"AUTH_REDIS_HOSTS" required:"true"`
	AuthRedisUserName string `env:"AUTH_REDIS_USERNAME" required:"true"`
	AuthRedisPassword string `env:"AUTH_REDIS_PASSWORD" required:"true"`
	AuthRedisPrefix   string `env:"AUTH_REDIS_PREFIX" required:"true"`

	AccountCookieName string `env:"ACCOUNT_COOKIE_NAME" required:"true"`
	ClusterCookieName string `env:"CLUSTER_COOKIE_NAME" required:"true"`

	KafkaBrokers  string `env:"KAFKA_BROKERS" required:"true"`
	KafkaUsername string `env:"KAFKA_USERNAME" required:"true"`
	KafkaPassword string `env:"KAFKA_PASSWORD" required:"true"`

	KafkaStatusUpdatesTopic string `env:"KAFKA_STATUS_UPDATES_TOPIC" required:"true"`
	KafkaErrorOnApplyTopic  string `env:"KAFKA_ERROR_ON_APPLY_TOPIC" required:"true"`
	KafkaConsumerGroupId    string `env:"KAFKA_CONSUMER_GROUP_ID" required:"true"`

	IAMGrpcAddr   string `env:"IAM_GRPC_ADDR" required:"true"`
	InfraGrpcAddr string `env:"INFRA_GRPC_ADDR" required:"true"`

	DefaultProjectWorkspaceName string `env:"DEFAULT_PROJECT_WORKSPACE_NAME" required:"true"`

	MsvcTemplateFilePath string `env:"MSVC_TEMPLATE_FILE_PATH" required:"true"`

	LokiServerHttpAddr string `env:"LOKI_SERVER_HTTP_ADDR" required:"true"`
	PromHttpAddr       string `env:"PROM_HTTP_ADDR" required:"true"`

	// AggregatedImagePullSecretName string `env:"AGGREGATED_IMAGE_PULL_SECRET_NAME" required:"true"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, err
	}
	return &e, nil
}
