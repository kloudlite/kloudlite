package env

import (
	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
)

type Env struct {
	InfraDbUri  string `env:"MONGO_DB_URI" required:"true"`
	InfraDbName string `env:"MONGO_DB_NAME" required:"true"`

	HttpPort     uint16 `env:"HTTP_PORT" required:"true"`
	GrpcPort     uint16 `env:"GRPC_PORT" required:"true"`
	CookieDomain string `env:"COOKIE_DOMAIN" required:"true"`

	NatsURL    string `env:"NATS_URL" required:"true"`
	NatsStream string `env:"NATS_STREAM" required:"true"`

	AccountCookieName       string `env:"ACCOUNT_COOKIE_NAME" required:"true"`
	ProviderSecretNamespace string `env:"PROVIDER_SECRET_NAMESPACE" required:"true"`

	IAMGrpcAddr      string `env:"IAM_GRPC_ADDR" required:"true"`
	AccountsGrpcAddr string `env:"ACCOUNTS_GRPC_ADDR" required:"true"`

	MessageOfficeInternalGrpcAddr string `env:"MESSAGE_OFFICE_INTERNAL_GRPC_ADDR" required:"true"`

	AWSCfParamTrustedARN           string `env:"AWS_CF_PARAM_TRUSTED_ARN" required:"true"`
	AWSCfStackNamePrefix           string `env:"AWS_CF_STACK_NAME_PREFIX" required:"true"`
	AWSCfRoleNamePrefix            string `env:"AWS_CF_ROLE_NAME_PREFIX" required:"true"`
	AWSCfInstanceProfileNamePrefix string `env:"AWS_CF_INSTANCE_PROFILE_NAME_PREFIX" required:"true"`
	AWSCfStackS3URL                string `env:"AWS_CF_STACK_S3_URL" required:"true"`

	AWSAccessKey string `env:"AWS_ACCESS_KEY" required:"true"`
	AWSSecretKey string `env:"AWS_SECRET_KEY" required:"true"`

	PublicDNSHostSuffix string `env:"PUBLIC_DNS_HOST_SUFFIX" required:"true"`
	SessionKVBucket     string `env:"SESSION_KV_BUCKET" required:"true"`
	IsDev               bool
	KubernetesApiProxy  string `env:"KUBERNETES_API_PROXY"`

	MsvcTemplateFilePath string `env:"MSVC_TEMPLATE_FILE_PATH" required:"true"`

	DeviceNamespace string `env:"DEVICE_NAMESPACE" default:"infra-devices"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, errors.NewE(err)
	}
	return &ev, nil
}
