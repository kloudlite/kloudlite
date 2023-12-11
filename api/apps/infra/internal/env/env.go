package env

import "github.com/codingconcepts/env"

type Env struct {
	InfraDbUri  string `env:"INFRA_DB_URI" required:"true"`
	InfraDbName string `env:"INFRA_DB_NAME" required:"true"`

	HttpPort     uint16 `env:"HTTP_PORT" required:"true"`
	GrpcPort     uint16 `env:"GRPC_PORT" required:"true"`
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	AuthRedisHosts    string `env:"AUTH_REDIS_HOSTS" required:"true"`
	AuthRedisUserName string `env:"AUTH_REDIS_USERNAME" required:"true"`
	AuthRedisPassword string `env:"AUTH_REDIS_PASSWORD" required:"true"`
	AuthRedisPrefix   string `env:"AUTH_REDIS_PREFIX" required:"true"`

	NatsURL    string `env:"NATS_URL" required:"true"`
	NatsStream string `env:"NATS_STREAM" required:"true"`

	AccountCookieName       string `env:"ACCOUNT_COOKIE_NAME" required:"true"`
	ProviderSecretNamespace string `env:"PROVIDER_SECRET_NAMESPACE" required:"true"`

	IAMGrpcAddr      string `env:"IAM_GRPC_ADDR" required:"true"`
	AccountsGrpcAddr string `env:"ACCOUNTS_GRPC_ADDR" required:"true"`

	MessageOfficeInternalGrpcAddr string `env:"MESSAGE_OFFICE_INTERNAL_GRPC_ADDR" required:"true"`

	VPNDevicesMaxOffset   int64 `env:"VPN_DEVICES_MAX_OFFSET" required:"true"`
	VPNDevicesOffsetStart int   `env:"VPN_DEVICES_OFFSET_START" required:"true"`

	AWSCfParamTrustedARN           string `env:"AWS_CF_PARAM_TRUSTED_ARN" required:"true"`
	AWSCfStackNamePrefix           string `env:"AWS_CF_STACK_NAME_PREFIX" required:"true"`
	AWSCfRoleNamePrefix            string `env:"AWS_CF_ROLE_NAME_PREFIX" required:"true"`
	AWSCfInstanceProfileNamePrefix string `env:"AWS_CF_INSTANCE_PROFILE_NAME_PREFIX" required:"true"`
	AWSCfStackS3URL                string `env:"AWS_CF_STACK_S3_URL" required:"true"`

	AWSAccessKey string `env:"AWS_ACCESS_KEY" required:"true"`
	AWSSecretKey string `env:"AWS_SECRET_KEY" required:"true"`

	PublicDNSHostSuffix string `env:"PUBLIC_DNS_HOST_SUFFIX" required:"true"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
