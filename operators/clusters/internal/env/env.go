package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`

	CloudflareApiToken string `env:"CLOUDFLARE_API_TOKEN" required:"true"`
	CloudflareZoneId   string `env:"CLOUDFLARE_ZONE_ID" required:"true"`

	KlS3BucketName   string `env:"KL_S3_BUCKET_NAME" required:"true"`
	KlS3BucketRegion string `env:"KL_S3_BUCKET_REGION" required:"true"`

	KlAwsAccessKey string `env:"KL_AWS_ACCESS_KEY" required:"true"`
	KlAwsSecretKey string `env:"KL_AWS_SECRET_KEY" required:"true"`

	MessageOfficeGRPCAddr string `env:"MESSAGE_OFFICE_GRPC_ADDR" required:"true"`

	IACJobImage string `env:"IAC_JOB_IMAGE" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}
