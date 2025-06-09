package kubectl

import (
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ControllerClient interface {
	client.Client
}

func NewClientWithScheme(restCfg *rest.Config, scheme *runtime.Scheme) (ControllerClient, error) {
	clientgoscheme.AddToScheme(scheme)
	cli, err := client.New(restCfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	return ControllerClient(cli), nil
}
