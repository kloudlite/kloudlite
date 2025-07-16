package env

import (
	"io"
	"os"

	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
	"sigs.k8s.io/yaml"
)

type AccountsEnv struct {
	KubernetesApiProxy              string `env:"ACCOUNTS__KUBERNETES_API_PROXY"`
	AvailableKloudliteRegionsConfig string `env:"ACCOUNTS__AVAILABLE_KLOUDLITE_REGIONS_CONFIG" required:"false"`
	AvailableKloudliteRegions       []AvailableKloudliteRegion
}

type AvailableKloudliteRegion struct {
	ID            string `json:"id"`
	DisplayName   string `json:"displayName"`
	Region        string `json:"region"`
	CloudProvider string `json:"cloudProvider"`
	Kubeconfig    string `json:"kubeconfig"`
	PublicDNSHost string `json:"publicDNSHost"`
}

func LoadEnv() (*AccountsEnv, error) {
	var ev AccountsEnv
	if err := env.Set(&ev); err != nil {
		return nil, errors.NewE(err)
	}

	if ev.AvailableKloudliteRegionsConfig != "" {
		f, err := os.Open(ev.AvailableKloudliteRegionsConfig)
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(b, &ev.AvailableKloudliteRegions); err != nil {
			return nil, err
		}
	}

	return &ev, nil
}
