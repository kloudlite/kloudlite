package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	*kubernetes.Clientset
}

func NewClient(cfg *rest.Config) (*Client, error) {
	kcli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{Clientset: kcli}, nil
}
