package env

import "github.com/codingconcepts/env"

type Env struct {
	InfraDbUri  string `env:"INFRA_DB_URI" required:"true"`
	InfraDbName string `env:"INFRA_DB_NAME" required:"true"`

	HttpPort     uint16 `env:"HTTP_PORT" required:"true"`
	GrpcPort     uint16 `env:"GRPC_PORT" required:"true"`
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	AuthRedisHosts    string `env:"AUTH_REDIS_HOSTS" required:"true"`
	AuthRedisUserName string `env:"AUTH_REDIS_USER_NAME" required:"true"`
	AuthRedisPassword string `env:"AUTH_REDIS_PASSWORD" required:"true"`
	AuthRedisPrefix   string `env:"AUTH_REDIS_PREFIX" required:"true"`

	KafkaBrokers         string `env:"KAFKA_BROKERS" required:"true"`
	KafkaUsername        string `env:"KAFKA_USERNAME" required:"true"`
	KafkaPassword        string `env:"KAFKA_PASSWORD" required:"true"`
	KafkaConsumerGroupId string `env:"KAFKA_CONSUMER_GROUP_ID" required:"true"`

	KafkaTopicSendMessagesToTargetWaitQueue string `env:"KAFKA_TOPIC_SEND_MESSAGES_TO_TARGET_WAIT_QUEUE" required:"true"`
	KafkaTopicInfraUpdates                  string `env:"KAFKA_TOPIC_INFRA_UPDATES" required:"true"`

	AccountCookieName       string `env:"ACCOUNT_COOKIE_NAME" required:"true"`
	ProviderSecretNamespace string `env:"PROVIDER_SECRET_NAMESPACE" required:"true"`

	// KloudliteReservedNamespace string `env:"KLOUDLITE_RESERVED_NAMESPACE" required:"true"`

	IAMGrpcAddr      string `env:"IAM_GRPC_ADDR" required:"true"`
	AccountsGrpcAddr string `env:"ACCOUNTS_GRPC_ADDR" required:"true"`

	MessageOfficeInternalGrpcAddr string `env:"MESSAGE_OFFICE_INTERNAL_GRPC_ADDR" required:"true"`

	VPNDevicesMaxOffset   int64 `env:"VPN_DEVICES_MAX_OFFSET" required:"true"`
	VPNDevicesOffsetStart int   `env:"VPN_DEVICES_OFFSET_START" required:"true"`

	AWSAssumeTenantRoleFormatString string `env:"AWS_ASSUME_TENANT_ROLE_FORMAT_STRING" required:"true"`

	AWSCloudformationParamExternalId string `env:"AWS_CLOUDFORMATION_PARAM_EXTERNAL_ID" required:"true"`
	AWSCloudformationParamTrustedARN string `env:"AWS_CLOUDFORMATION_PARAM_TRUSTED_ARN" required:"true"`
	AWSCloudformationStackName       string `env:"AWS_CLOUDFORMATION_STACK_NAME" required:"true"`
	AWSCloudformationStackS3URL      string `env:"AWS_CLOUDFORMATION_STACK_S3_URL" required:"true"`

	AWSAccessKey string `env:"AWS_ACCESS_KEY" required:"true"`
	AWSSecretKey string `env:"AWS_SECRET_KEY" required:"true"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
