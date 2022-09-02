package k8s

import (
	"flag"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"kloudlite.io/pkg/errors"
)

func getKubeConfig(isDev bool) (cfg *rest.Config, e error) {
	if !isDev {
		cfg, e = rest.InClusterConfig()
		if e != nil {
			return nil, errors.NewEf(e, "could not read incluster kubeconfig")
		}
		return
	}

	cfgPath, ok := os.LookupEnv("KUBECONFIG")
	if !ok {
		return nil, errors.New("KUBECONFIG env var is not present")
	}

	return clientcmd.BuildConfigFromFlags("", cfgPath)
}

func NewClient() (*kubernetes.Clientset, error) {
	isDev := flag.Bool("dev", false, "is development")
	cfg, e := getKubeConfig(*isDev)
	if e != nil {
		return nil, errors.NewEf(e, "could not retrieve kubeconfig")
	}
	return kubernetes.NewForConfig(cfg)
}
