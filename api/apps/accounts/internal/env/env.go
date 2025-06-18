package env

import (
	"io"
	"os"

	"github.com/codingconcepts/env"
	"github.com/kloudlite/api/pkg/errors"
	"sigs.k8s.io/yaml"
)

type Env struct {
	GrpcPort uint16 `env:"GRPC_PORT" required:"true"`

	DBName string `env:"MONGO_DB_NAME" required:"true"`
	DBUrl  string `env:"MONGO_URI" required:"true"`

	IamGrpcAddr   string `env:"IAM_GRPC_ADDR" required:"true"`
	CommsGrpcAddr string `env:"COMMS_GRPC_ADDR" required:"true"`
	// ContainerRegistryGrpcAddr string `env:"CONTAINER_REGISTRY_GRPC_ADDR" required:"true"`
	ConsoleGrpcAddr string `env:"CONSOLE_GRPC_ADDR" required:"true"`
	AuthGrpcAddr    string `env:"AUTH_GRPC_ADDR" required:"true"`

	SessionKVBucket string `env:"SESSION_KV_BUCKET" required:"true"`
	NatsURL         string `env:"NATS_URL" required:"true"`

	IsDev              bool
	KubernetesApiProxy string `env:"KUBERNETES_API_PROXY"`

	AvailableKloudliteRegionsConfig string `env:"AVAILABLE_KLOUDLITE_REGIONS_CONFIG" required:"false"`
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
