package env

import "github.com/codingconcepts/env"

type Env struct {
	CloudProvider string `env:"CLOUD_PROVIDER" required:"true"`
	Action        string `env:"ACTION" required:"true"`

	NodeConfig     string `env:"NODE_CONFIG" required:"true"`
	ProviderConfig string `env:"PROVIDER_CONFIG" required:"true"`

	AWSProviderConfig   string `env:"AWS_PROVIDER_CONFIG"`
	GCPProviderConfig   string `env:"GCP_PROVIDER_CONFIG"`
	AzureProviderConfig string `env:"AZURE_PROVIDER_CONFIG"`
	DoProviderConfig    string `env:"DO_PROVIDER_CONFIG"`

	AgentHelmValues     string `env:"AGENT_HELM_VALUES"`
	OperatorsHelmValues string `env:"OPERATORS_HELM_VALUES"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Set(&e); err != nil {
		return nil, err
	}
	return &e, nil
}
