package env

import (
	"github.com/codingconcepts/env"
)

type Env struct {
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

	AWSVpcId            string `env:"AWS_VPC_ID"`
	AWSVpcPublicSubnets string `env:"AWS_VPC_PUBLIC_SUBNETS"`
}

func GetEnvOrDie() *Env {
	var ev Env
	if err := env.Set(&ev); err != nil {
		panic(err)
	}

	return &ev
}
