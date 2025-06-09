package env

import (
	"encoding/json"
	"os"

	"github.com/codingconcepts/env"
	corev1 "k8s.io/api/core/v1"
)

type realEnv struct {
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

type Env struct {
	realEnv
	IACJobTolerations  []corev1.Toleration `env:"IAC_JOB_TOLERATIONS" required:"false"`
	IACJobNodeSelector map[string]string   `env:"IAC_JOB_NODE_SELECTOR" required:"false"`
}

func GetEnvOrDie() *Env {
	var rev realEnv
	if err := env.Set(&rev); err != nil {
		panic(err)
	}

	ev := &Env{
		realEnv: rev,
	}

	ev.IACJobTolerations = []corev1.Toleration{}
	if v, ok := os.LookupEnv("IAC_JOB_TOLERATIONS"); ok {
		if err := json.Unmarshal([]byte(v), &ev.IACJobTolerations); err != nil {
			panic(err)
		}
	}

	ev.IACJobNodeSelector = map[string]string{}
	if v, ok := os.LookupEnv("IAC_JOB_NODE_SELECTOR"); ok {
		if err := json.Unmarshal([]byte(v), &ev.IACJobNodeSelector); err != nil {
			panic(err)
		}
	}

	return ev
}
