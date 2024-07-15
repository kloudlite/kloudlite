package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
	baseEnv
	kloudliteNodepoolEnv
}

type baseEnv struct {
	// common
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`

	EnableNodepools bool `env:"ENABLE_NODEPOOLS" required:"true"`
}

type kloudliteNodepoolEnv struct {
	JobsNamespace string `env:"JOBS_NAMESPACE" default:"kloudlite-jobs"`

	AccountName string `env:"ACCOUNT_NAME" required:"true"` // required only for labelling nodepool nodes with it
	ClusterName string `env:"CLUSTER_NAME" required:"true"` // required only for labelling nodepool nodes with it

	K3sJoinToken        string `env:"K3S_JOIN_TOKEN" required:"true"`
	K3sServerPublicHost string `env:"K3S_SERVER_PUBLIC_HOST" required:"true"`

	TFStateSecretNamespace string `env:"TF_STATE_SECRET_NAMESPACE" required:"true" default:"kloudlite"`
	IACJobImage            string `env:"IAC_JOB_IMAGE" required:"true"`

	// for, `k3s-runner`, and `k3s` binary on the to be created VM.
	KloudliteRelease string `env:"KLOUDLITE_RELEASE" required:"true"`
}

func GetEnvOrDie() *Env {
	var ev Env

	var bev baseEnv
	if err := env.Set(&bev); err != nil {
		panic(err)
	}

	ev.baseEnv = bev

	if bev.EnableNodepools {
		var npenv kloudliteNodepoolEnv
		if err := env.Set(&npenv); err != nil {
			panic(err)
		}
		ev.kloudliteNodepoolEnv = npenv
	}

	return &ev
}
