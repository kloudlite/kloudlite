package env

type Env struct {
	// KafkaInfraTopic      string `env:"KAFKA_INFRA_TOPIC" required:"true"`
	ManagedTemplatesPath string `env:"MANAGED_TEMPLATES_PATH" required:"true"`
	InventoryPath        string `env:"INVENTORY_PATH" required:"true"`

	WorkloadStatusTopic string `env:"KAFKA_WORKLOAD_STATUS_TOPIC"`
	WorkloadApplyTopic  string `env:"KAFKA_WORKLOAD_TOPIC"`

	KafkaConsumerGroupId string `env:"KAFKA_GROUP_ID"`
	CookieDomain         string `env:"COOKIE_DOMAIN"`

	// ResponseTopic string `env:"KAFKA_WORKLOAD_RESP_TOPIC"`

	LokiServerUrl  string `env:"LOKI_URL" required:"true"`
	LogServerPort  uint64 `env:"LOG_SERVER_PORT" required:"true"`
	JSEvalService  string `env:"JSEVAL_SERVICE"`
	IAMService     string `env:"IAM_SERVICE"`
	CIService      string `env:"CI_SERVICE" required:"true"`
	AuthService    string `env:"AUTH_SERVICE" required:"true"`
	FinanceService string `env:"FINANCE_SERVICE" required:"true"`

	MongoUri      string `env:"MONGO_URI" required:"true"`
	RedisHosts    string `env:"REDIS_HOSTS" required:"true"`
	RedisUserName string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisPrefix   string `env:"REDIS_PREFIX"`

	AuthRedisHosts    string `env:"REDIS_AUTH_HOSTS" required:"true"`
	AuthRedisUserName string `env:"REDIS_AUTH_USERNAME"`
	AuthRedisPassword string `env:"REDIS_AUTH_PASSWORD"`
	AuthRedisPrefix   string `env:"REDIS_AUTH_PREFIX" required:"true"`

	MongoDbName  string `env:"MONGO_DB_NAME" required:"true"`
	KafkaBrokers string `env:"KAFKA_BOOTSTRAP_SERVERS" required:"true"`
	Port         uint16 `env:"PORT" required:"true"`
	IsDev        bool   `env:"DEV" default:"false" required:"true"`

	GrpcPort    uint16 `env:"GRPC_PORT" required:"true"`
	NotifierUrl string `env:"NOTIFIER_URL" required:"true"`
}
