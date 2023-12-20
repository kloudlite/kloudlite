package k8s

import (
	"encoding/json"
	"github.com/kloudlite/api/pkg/errors"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func RestInclusterConfig() (*rest.Config, error) {
	return rest.InClusterConfig()
}

func RestConfigFromEnv(envVar string) (*rest.Config, error) {
	kcfgPath, ok := os.LookupEnv(envVar)
	if !ok {
		return nil, errors.Newf("KUBECONFIG env variable is not set")
	}

	return clientcmd.BuildConfigFromKubeconfigGetter("", func() (*api.Config, error) {
		b, err := os.ReadFile(kcfgPath)
		if err != nil {
			return nil, err
		}

		var kubeconfig api.Config
		if err := json.Unmarshal(b, &kubeconfig); err != nil {
			return nil, err
		}

		return &kubeconfig, nil
	})
}

func RestConfigFromKubeConfig(b []byte) (*rest.Config, error) {
	return clientcmd.BuildConfigFromKubeconfigGetter("", func() (*api.Config, error) {
		var kubeconfig api.Config
		if err := json.Unmarshal(b, &kubeconfig); err != nil {
			return nil, err
		}

		return &kubeconfig, nil
	})
}

