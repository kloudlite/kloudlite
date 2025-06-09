package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`

	K3sParamsSecretRef string `env:"K3S_PARAMS_SECRET_REF" required:"true"`

	IACJobsNamespace string `env:"IAC_JOBS_NAMESPACE" required:"true"`
	IACJobImage      string `env:"IAC_JOB_IMAGE" required:"true"`

	TFStateSecretNamespace string `env:"TF_STATE_SECRET_NAMESPACE" required:"true" default:"kloudlite"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}

	return &ev
}
