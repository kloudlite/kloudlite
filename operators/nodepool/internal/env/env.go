package env

import (
	"strings"

	"github.com/codingconcepts/env"
)

type Env struct {
	// common
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES"`

	CloudProviderName   string `env:"CLOUD_PROVIDER_NAME" required:"true"`
	CloudProviderRegion string `env:"CLOUD_PROVIDER_REGION" required:"true"`

	JobsNamespace string `env:"JOBS_NAMESPACE" default:"kloudlite-jobs"`

	AccountName string `env:"ACCOUNT_NAME" required:"true"` // required only for labelling nodepool nodes with it
	ClusterName string `env:"CLUSTER_NAME" required:"true"` // required only for labelling nodepool nodes with it

	K3sJoinToken        string `env:"K3S_JOIN_TOKEN" required:"true"`
	K3sServerPublicHost string `env:"K3S_SERVER_PUBLIC_HOST" required:"true"`

	TFStateSecretNamespace string `env:"TF_STATE_SECRET_NAMESPACE" required:"true" default:"kloudlite"`
	IACJobImage            string `env:"IAC_JOB_IMAGE" required:"true"`

	EnableNodepools bool `env:"ENABLE_NODEPOOLS" required:"true"`

	// for, `k3s-runner`, and `k3s` binary on the to be created VM.
	KloudliteRelease string `env:"KLOUDLITE_RELEASE"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}

	if ev.EnableNodepools {
		// if strings.TrimSpace(ev.AWSVpcId) == "" {
		// 	panic("env-var AWS_VPC_ID must be set, when ENABLE_NODEPOOLS is true")
		// }
		//
		// if strings.TrimSpace(ev.AWSVpcPublicSubnets) == "" {
		// 	panic("env-var AWS_VPC_PUBLIC_SUBNETS must be set, when ENABLE_NODEPOOLS is true")
		// }

		if strings.TrimSpace(ev.KloudliteRelease) == "" {
			panic("env-var KLOUDLITE_RELEASE must be set, when ENABLE_NODEPOOLS is true")
		}
	}

	return &ev
}
