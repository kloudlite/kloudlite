package kubeapi

import (
	"context"
	"encoding/json"
	"io"
	v1 "k8s.io/api/core/v1"
	"net/http"
)

type Client struct {
	Address string `env:"KUBE_API_ADDRESS"`
}

func (c *Client) GetSecret(ctx context.Context, namespace, name string) (*v1.Secret, error) {
	secret := v1.Secret{}
	get, err := http.Get(c.Address + "/api/v1/namespaces/" + namespace + "/secrets/" + name)
	if err != nil {
		return nil, err
	}
	defer get.Body.Close()
	all, err := io.ReadAll(get.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(all, &secret); err != nil {
		return nil, err
	}
	return &secret, nil
}

func NewClient(addr string) *Client {
	return &Client{
		Address: addr,
	}
}
