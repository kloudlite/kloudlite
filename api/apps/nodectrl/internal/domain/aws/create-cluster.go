package aws

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	guuid "github.com/google/uuid"
	"gopkg.in/yaml.v2"

	"kloudlite.io/apps/nodectrl/internal/domain/common"
	"kloudlite.io/apps/nodectrl/internal/domain/utils"
)

// CreateCluster implements common.ProviderClient
func (a AwsClient) CreateCluster(ctx context.Context) error {
	/*
		create node
		check for rediness
		install k3s
		check for rediness
		install maaster
	*/
	if err := a.ensureForMasters(); err != nil {
		return err
	}

	if err := a.SetupSSH(); err != nil {
		return err
	}
	defer a.saveForSure()
	sshPath := path.Join("/tmp/ssh", a.accountName)

	if err := a.NewNode(ctx); err != nil {
		return err
	}

	ip, err := utils.GetOutput(path.Join(utils.Workdir, *a.node.NodeName), "node-ip")
	if err != nil {
		return err
	}

	count := 0

	for {
		if e := utils.ExecCmd(
			fmt.Sprintf("ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s root@%s ls",
				fmt.Sprintf("%v/access", sshPath),
				string(ip),
			),
			"checking is node is ready"); e == nil {
			break
		}

		count++
		if count > 24 {
			return fmt.Errorf("node is not ready even after 6 minutes")
		}
		time.Sleep(time.Second * 5)
	}

	masterToken := guuid.New()

	// install k3s
	cmd := fmt.Sprintf(
		"ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s/access root@%s sudo sh /tmp/k3s-install.sh server --token=%s --node-external-ip %s --flannel-backend wireguard-native --flannel-external-ip --disable traefik --node-name=%s --cluster-init",
		sshPath,
		string(ip),
		masterToken.String(),
		string(ip),
		*a.node.NodeName,
	)

	if err := utils.ExecCmd(cmd, "installing k3s"); err != nil {
		return err
	}
	// needed to fetch kubeconfig

	configOut, err := utils.ExecCmdWithOutput(fmt.Sprintf("ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s/access root@%s cat /etc/rancher/k3s/k3s.yaml", sshPath, string(ip)), "fetching kubeconfig from the cluster")
	if err != nil {
		return err
	}

	var kubeconfig common.KubeConfigType
	if err := yaml.Unmarshal(configOut, &kubeconfig); err != nil {
		return err
	}

	for i := range kubeconfig.Clusters {
		kubeconfig.Clusters[i].Cluster.Server = fmt.Sprintf("https://%s:6443", string(ip))
	}

	kc, err := yaml.Marshal(kubeconfig)
	if err != nil {
		return err
	}

	tokenOut, err := utils.ExecCmdWithOutput(fmt.Sprintf("ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s/access root@%s cat /var/lib/rancher/k3s/server/node-token", sshPath, string(ip)), "fetching node token from the cluster")
	if err != nil {
		return err
	}

	st := TokenAndKubeconfig{
		Token:       string(tokenOut),
		Kubeconfig:  string(kc),
		ServerIp:    string(ip),
		MasterToken: masterToken.String(),
	}

	b, err := yaml.Marshal(st)
	if err != nil {
		return err
	}

	tokenPath := path.Join(sshPath, "config.yaml")

	if err := os.WriteFile(tokenPath, b, os.ModePerm); err != nil {
		return err
	}

	if err := a.awsS3Client.UploadFile(tokenPath, fmt.Sprintf("%s-config.yaml", a.accountName)); err != nil {
		return err
	}

	return err
}

func parseValues(a AwsClient, sshPath string) (map[string]any, error) {
	returnError := func(errorFor string) (map[string]any, error) {
		return nil, fmt.Errorf("required value %q not provided", errorFor)
	}

	values := map[string]any{}

	values["access_key"] = a.accessKey
	values["secret_key"] = a.accessSecret

	if a.node.Region == nil {
		return returnError("region")
	}
	values["region"] = *a.node.Region

	values["keys_path"] = sshPath

	if a.node.NodeName == nil {
		return returnError("nodename")
	}
	values["node_name"] = *a.node.NodeName

	if a.node.ProvisionMode != "spot" {
		if a.node.OnDemandSpecs == nil {
			return returnError("onDemandSpecs")
		}
		values["instance_type"] = a.node.OnDemandSpecs.InstanceType

	}

	if a.node.ProvisionMode == "spot" {
		values["cpu_min"] = a.node.SpotSpecs.CpuMin
		values["cpu_max"] = a.node.SpotSpecs.CpuMax

		values["mem_min"] = a.node.SpotSpecs.CpuMin
		values["mem_max"] = a.node.SpotSpecs.MemMax
	}

	if a.node.ImageId != nil {
		values["ami"] = *a.node.ImageId
	}

	return values, nil
}
