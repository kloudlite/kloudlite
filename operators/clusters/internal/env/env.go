package env

import (
	"time"

	"github.com/codingconcepts/env"
)

type Env struct {
	ReconcilePeriod         time.Duration `env:"RECONCILE_PERIOD"`
	MaxConcurrentReconciles int           `env:"MAX_CONCURRENT_RECONCILES"`

	// JobNamespace string `env:"JOB_NAMESPACE" required:"true"`

	// RunMode string `env:"RUN_MODE" required:"true"` // enum[ platform,target ]

	CloudflareApiToken string `env:"CLOUDFLARE_API_TOKEN" required:"true"`
	CloudflareZoneId   string `env:"CLOUDFLARE_ZONE_ID" required:"true"`

	KlS3BucketName   string `env:"KL_S3_BUCKET_NAME" required:"true"`
	KlS3BucketRegion string `env:"KL_S3_BUCKET_REGION" required:"true"`

	KlAwsAccessKey       string `env:"KL_AWS_ACCESS_KEY" required:"true"`
	KlAwsSecretAccessKey string `env:"KL_AWS_SECRET_ACCESS_KEY" required:"true"`
}

type TargetEnv struct {
	AccountName   string `env:"ACCOUNT_NAME" required:"true"`
	AccountId     string `env:"AWS_ACCOUNT_ID"`
	AccessKey     string `env:"ACCESS_KEY" required:"true"`
	AccessSecret  string `env:"ACCESS_SECRET" required:"true"`
	CloudProvider string `env:"CLOUD_PROVIDER" required:"true"`
}

type PlatformEnv struct {
	// example|default: dns.khost.dev
	DnsHostedZone string `env:"DNS_HOSTED_ZONE" required:"true"`
	AccessKey     string `env:"ACCESS_KEY" required:"true"`
	AccessSecret  string `env:"ACCESS_SECRET" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}
	return &ev
}

// func GetTargetEnvOrDie() *TargetEnv {
// 	var ev TargetEnv
// 	if err := env.Set(&ev); err != nil {
// 		panic(err)
// 	}
// 	return &ev
// }
//
// func GetPlatformEnvOrDie() *PlatformEnv {
// 	var ev PlatformEnv
// 	if err := env.Set(&ev); err != nil {
// 		panic(err)
// 	}
// 	return &ev
// }
