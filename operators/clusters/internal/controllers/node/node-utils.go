package node

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
)

func getProviderConfig() (string, error) {
	pd := CommonProviderData{
		TfTemplates: tfTemplates,
		Labels:      map[string]string{},
		Taints:      []string{},
		SSHPath:     "",
	}
	out, err := yaml.Marshal(pd)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out), nil
}

func (r *Reconciler) getNodeConfig(np *clustersv1.NodePool, obj *clustersv1.Node) (string, error) {
	switch r.Env.CloudProvider {
	case "aws":
		var awsNode AWSNode
		if err := json.Unmarshal([]byte(np.Spec.NodeConfig), &awsNode); err != nil {
			return "", err
		}

		awsNode.NodeId = obj.Name

		awsbyte, err := yaml.Marshal(awsNode)
		if err != nil {
			return "", err
		}
		// return awsbyte, nil

		return base64.StdEncoding.EncodeToString(awsbyte), nil

	case "do", "azure", "gcp":
		panic("unimplemented")
	default:
		return "", fmt.Errorf("this type of cloud provider not supported for now")
	}
}

func (r *Reconciler) getSpecificProvierConfig() (string, error) {
	switch r.Env.CloudProvider {
	case "aws":
		out, err := json.Marshal(AwsProviderConfig{
			AccessKey:    r.Env.AccessKey,
			AccessSecret: r.Env.AccessSecret,
			AccountName:  r.Env.AccountName,
		})
		if err != nil {
			return "", err
		}

		return base64.StdEncoding.EncodeToString(out), nil
	default:
		return "", fmt.Errorf("cloud provider %s not supported for now", r.Env.CloudProvider)
	}
}

func getAction(obj *clustersv1.Node) string {
	switch obj.Spec.NodeType {
	case "worker":
		return "add-worker"
	case "master":
		return "add-master"
	case "cluster":
		return "create-cluster"
	default:
		return "unknown"
	}
}
