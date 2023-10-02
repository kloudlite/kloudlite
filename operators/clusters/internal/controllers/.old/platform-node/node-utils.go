package platform_node

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operators/clusters/internal/controllers/target-node"
	fn "github.com/kloudlite/operator/pkg/functions"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getProviderConfig() (string, error) {
	out, err := yaml.Marshal(target_node.CommonProviderData{
		TfTemplates: tfTemplates,
		Labels:      map[string]string{},
		Taints:      []string{},
		SSHPath:     "",
	})
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out), nil
}

func (r *Reconciler) getNodeConfig(cl *clustersv1.Cluster, obj *clustersv1.Node) (string, error) {
	switch cl.Spec.CloudProvider {
	case "aws":
		awsNode := target_node.AWSNodeConfig{
			NodeName: obj.Name,
			AWSNodeConfig: clustersv1.AWSNodeConfig{
				OnDemandSpecs: &clustersv1.OnDemandSpecs{
					InstanceType: "c6a.xlarge",
				},
				Region:        &cl.Spec.Region,
				ProvisionMode: "on-demand",
				VPC:           cl.Spec.VPC,
			},
		}

		awsbyte, err := yaml.Marshal(awsNode)
		if err != nil {
			return "", err
		}

		return base64.StdEncoding.EncodeToString(awsbyte), nil

	case "do", "azure", "gcp":
		panic("unimplemented")
	default:
		return "", fmt.Errorf("this type of cloud provider not supported for now")
	}
}

func getSpecificProvierConfig(ctx context.Context, client client.Client, cl *clustersv1.Cluster) (string, error) {

	s, err := rApi.Get(ctx, client, fn.NN(cl.Spec.CredentialsRef.Namespace, cl.Spec.CredentialsRef.Name), &corev1.Secret{})
	if err != nil {
		return "", err
	}

	switch cl.Spec.CloudProvider {
	case "aws":
		out, err := json.Marshal(target_node.AwsProviderConfig{
			AccessKey:    string(s.Data["accessKey"]),
			AccessSecret: string(s.Data["accessSecret"]),
			AccountName:  cl.Spec.AccountName,
		})
		if err != nil {
			return "", err
		}

		return base64.StdEncoding.EncodeToString(out), nil
	default:
		return "", fmt.Errorf("cloud provider %s not supported for now", cl.Spec.CloudProvider)
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
