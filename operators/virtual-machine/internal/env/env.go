package env

import (
	"encoding/json"
	"os"

	"github.com/codingconcepts/env"
	corev1 "k8s.io/api/core/v1"
)

type pEnv struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES" required:"true"`

	IACJobImage string `env:"IAC_JOB_IMAGE" required:"true"`
}

type Env struct {
	pEnv
	IACJobTolerations  []corev1.Toleration `env:"IAC_JOB_TOLERATIONS" required:"true"`
	IACJobNodeSelector map[string]string   `env:"IAC_JOB_NODE_SELECTOR" required:"true"`
}

func LoadEnvOrDie() *Env {
	var pev pEnv
	if err := env.Set(&pev); err != nil {
		panic(err)
	}

	ev := &Env{
		pEnv: pev,
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
