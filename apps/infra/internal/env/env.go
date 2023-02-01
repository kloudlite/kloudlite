package env

type Env struct {
	InfraDbUri  string `env:"INFRA_DB_URI" required:"true"`
	InfraDbName string `env:"INFRA_DB_NAME" required:"true"`

	HttpPort     uint16 `env:"HTTP_PORT" required:"true"`
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	FinanceGrpcAddr string `env:"FINANCE_GRPC_ADDR" required:"true"`

	AuthRedisHosts    string `env:"AUTH_REDIS_HOSTS" required:"true"`
	AuthRedisUserName string `env:"AUTH_REDIS_USER_NAME" required:"true"`
	AuthRedisPassword string `env:"AUTH_REDIS_PASSWORD" required:"true"`
	AuthRedisPrefix   string `env:"AUTH_REDIS_PREFIX" required:"true"`
}
