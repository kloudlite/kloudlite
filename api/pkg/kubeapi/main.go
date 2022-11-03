package kubeapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/exec"

	v1 "k8s.io/api/core/v1"
)

type Client struct {
	KubeconfigPath string
}
type AccountNodeStatus struct {
	IsReady          bool               `json:"isReady"`
	Conditions       []metav1.Condition `json:"conditions,omitempty"`
	StatusConditions []metav1.Condition `json:"statusConditions,omitempty"`
	OpsConditions    []metav1.Condition `json:"opsConditions,omitempty"`
	Message          string             `json:"message,omitempty"`
}

type AccountNodeSpec struct {
	AccountRef  string `json:"accountRef,omitempty"`
	Region      string `json:"region"`
	EdgeRef     string `json:"edgeRef"`
	Provider    string `json:"provider"`
	ProviderRef string `json:"providerRef,omitempty"`
	Config      string `json:"config"`
	Pool        string `json:"pool"`
	Index       int    `json:"nodeIndex,omitempty"`
}

type AccountNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccountNodeSpec   `json:"spec,omitempty"`
	Status AccountNodeStatus `json:"status,omitempty"`
}

type AccountNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccountNode `json:"items"`
}

func (c *Client) GetAccountNodes(ctx context.Context, edgeId string) (*AccountNodeList, error) {
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd := exec.Command("kubectl", "get", "accountnodes", "-l", fmt.Sprintf("kloudlite.io/region:%s", edgeId), "-o", "json")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", c.KubeconfigPath))

	if err := cmd.Run(); err != nil {
		fmt.Println(stderr.String())
		return &AccountNodeList{}, nil
	}

	var nodeList AccountNodeList
	if err := json.Unmarshal(stdout.Bytes(), &nodeList); err != nil {
		return nil, err
	}

	return &nodeList, nil
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
