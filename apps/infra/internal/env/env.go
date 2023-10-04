package env

type Env struct {
	InfraDbUri  string `env:"INFRA_DB_URI" required:"true"`
	InfraDbName string `env:"INFRA_DB_NAME" required:"true"`

	HttpPort     uint16 `env:"HTTP_PORT" required:"true"`
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	AuthRedisHosts    string `env:"AUTH_REDIS_HOSTS" required:"true"`
	AuthRedisUserName string `env:"AUTH_REDIS_USER_NAME" required:"true"`
	AuthRedisPassword string `env:"AUTH_REDIS_PASSWORD" required:"true"`
	AuthRedisPrefix   string `env:"AUTH_REDIS_PREFIX" required:"true"`

	KafkaBrokers         string `env:"KAFKA_BROKERS" required:"true"`
	KafkaUsername        string `env:"KAFKA_USERNAME" required:"true"`
	KafkaPassword        string `env:"KAFKA_PASSWORD" required:"true"`
	KafkaConsumerGroupId string `env:"KAFKA_CONSUMER_GROUP_ID" required:"true"`

	KafkaTopicInfraUpdates string `env:"KAFKA_TOPIC_INFRA_UPDATES" required:"true"`
	// KafkaTopicByocHelmUpdates   string `env:"KAFKA_TOPIC_BYOC_HELM_UPDATES" required:"true"`
	KafkaTopicByocClientUpdates string `env:"KAFKA_TOPIC_BYOC_CLIENT_UPDATES" required:"true"`

	AccountCookieName       string `env:"ACCOUNT_COOKIE_NAME" required:"true"`
	ProviderSecretNamespace string `env:"PROVIDER_SECRET_NAMESPACE" required:"true"`

	// KloudliteReservedNamespace string `env:"KLOUDLITE_RESERVED_NAMESPACE" required:"true"`

	IAMGrpcAddr      string `env:"IAM_GRPC_ADDR" required:"true"`
	AccountsGrpcAddr string `env:"ACCOUNTS_GRPC_ADDR" required:"true"`

	MessageOfficeInternalGrpcAddr string `env:"MESSAGE_OFFICE_INTERNAL_GRPC_ADDR" required:"true"`
}
