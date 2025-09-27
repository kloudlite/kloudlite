package env

import (
	"io"
	"os"

	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
	"sigs.k8s.io/yaml"
)

type Env struct {
	KubernetesApiProxy              string `env:"ACCOUNTS__KUBERNETES_API_PROXY"`
	AvailableKloudliteRegionsConfig string `env:"ACCOUNTS__AVAILABLE_KLOUDLITE_REGIONS_CONFIG" required:"false"`
	AvailableKloudliteRegions       []AvailableKloudliteRegion
	WebURL                          string `env:"ACCOUNTS__WEB_URL" required:"true"`
	
	// Platform configuration
	PlatformOwnerEmail              string `env:"AUTH__PLATFORM_OWNER_EMAIL" required:"false"`
	
	// JWT Configuration (should match auth service)
	JWTSecret                       string `env:"AUTH__JWT_SECRET" required:"true"`
}

type AvailableKloudliteRegion struct {
	ID            string `json:"id"`
	DisplayName   string `json:"displayName"`
	Region        string `json:"region"`
	CloudProvider string `json:"cloudProvider"`
	Kubeconfig    string `json:"kubeconfig"`
	PublicDNSHost string `json:"publicDNSHost"`
}

func LoadEnv() (*Env, error) {
	var ev Env
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
