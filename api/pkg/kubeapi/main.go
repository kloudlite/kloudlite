package kubeapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	v1 "k8s.io/api/core/v1"
)

type Client struct {
	KubeconfigPath string
}

func (c *Client) GetSecret(ctx context.Context, namespace, name string) (*v1.Secret, error) {
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd := exec.Command("kubectl", "get", fmt.Sprintf("secret/%s", name), "-n", namespace, "-o", "json")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", c.KubeconfigPath))

	if err := cmd.Run(); err != nil {
		fmt.Println(stderr.String())
		return &v1.Secret{}, nil
	}

	var secret v1.Secret
	if err := json.Unmarshal(stdout.Bytes(), &secret); err != nil {
		return nil, err
	}

	return &secret, nil
}

func NewClientWithConfigPath(cfgPath string) *Client {
	return &Client{KubeconfigPath: cfgPath}
}
