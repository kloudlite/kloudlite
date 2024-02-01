package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`

	CloudflareApiToken string `env:"CLOUDFLARE_API_TOKEN" required:"true"`
	CloudflareZoneId   string `env:"CLOUDFLARE_ZONE_ID" required:"true"`

	KlAwsAccessKey string `env:"KL_AWS_ACCESS_KEY" required:"true"`
	KlAwsSecretKey string `env:"KL_AWS_SECRET_KEY" required:"true"`

	MessageOfficeGRPCAddr string `env:"MESSAGE_OFFICE_GRPC_ADDR" required:"true"`

	IACJobImage string `env:"IAC_JOB_IMAGE" required:"true"`

	NatsURL    string `env:"NATS_URL" required:"true"`
	NatsStream string `env:"NATS_STREAM" required:"true"`

	NatsClusterUpdateSubjectFormat string `env:"NATS_CLUSTER_UPDATE_SUBJECT_FORMAT" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
