package k3s

import (
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
)

type K3sClient interface {
	CreateClustersTeams(name string) error
	EnsureKloudliteNetwork() error
	StartAppInterceptService(ports []apiclient.AppPort, toStart bool) error
	EnsureImage(i string) error
	RestartWgProxyContainer() error
	RemoveAllIntercepts() error
}

type client struct {
	c    *dockerclient.Client
	apic apiclient.ApiClient
	fc   fileclient.FileClient
}

func NewClient() (K3sClient, error) {
	apiClient, err := apiclient.New()
	if err != nil {
		return nil, err
	}
	fc, err := fileclient.New()
	if err != nil {
		return nil, err
	}

	c, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &client{
		c:    c,
		apic: apiClient,
		fc:   fc,
	}, nil
}
