package kubeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"os/exec"
)

type AddrClientImpl struct {
	address string
}

type ConfigClientImpl struct {
	configPath string
}

type Client interface {
	GetServiceIp(ctx context.Context, namespace, name string) (string, error)
	GetSecret(ctx context.Context, namespace, name string) (*v1.Secret, error)
}

func (c *ConfigClientImpl) GetServiceIp(ctx context.Context, namespace, name string) (string, error) {
	command := exec.Command("kubectl", "get", "svc", fmt.Sprintf("%s/%s", namespace, name), "-o", "jsonpath={.spec.clusterIP}")
	output, err := command.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil

}

func (c *AddrClientImpl) GetServiceIp(ctx context.Context, namespace, name string) (string, error) {
	service := v1.Service{}
	get, err := http.Get(c.address + "/api/v1/namespaces/" + namespace + "/services/" + name)
	if err != nil {
		return "", err
	}
	defer get.Body.Close()
	all, err := io.ReadAll(get.Body)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(all, &service); err != nil {
		return "", err
	}
	return service.Spec.ClusterIP, nil
}

func (c *ConfigClientImpl) GetSecret(ctx context.Context, namespace, name string) (*v1.Secret, error) {
	secret := v1.Secret{}
	command := exec.Command("kubectl", "get", "svc", fmt.Sprintf("%s/%s", namespace, name), "-o", "json")
	output, err := command.Output()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(output, &secret); err != nil {
		return nil, err
	}
	return &secret, nil
}

func (c *AddrClientImpl) GetSecret(ctx context.Context, namespace, name string) (*v1.Secret, error) {
	secret := v1.Secret{}
	get, err := http.Get(c.address + "/api/v1/namespaces/" + namespace + "/secrets/" + name)
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

func NewClientWithConfigPath(configPath string) Client {
	return &ConfigClientImpl{
		configPath: configPath,
	}
}

func NewClient(addr string) Client {
	return &AddrClientImpl{
		address: addr,
	}
}
