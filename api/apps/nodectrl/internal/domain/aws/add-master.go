package aws

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"kloudlite.io/apps/nodectrl/internal/domain/utils"
)

func (a AwsClient) ensureForMasters() error {
	switch a.node.NodeType {
	case "spot":
		return fmt.Errorf("spot is not supported as a master")
	default:
		return nil
	}
}

// AddMaster implements common.ProviderClient.
func (a AwsClient) AddMaster(ctx context.Context) error {
	if err := a.ensureForMasters(); err != nil {
		return err
	}

	if err := a.ensurePaths(); err != nil {
		return err
	}

	// fetch token
	sshPath := path.Join("/tmp/ssh", a.accountName)

	tokenFileName := fmt.Sprintf("%s-config.yaml", a.accountName)

	if err := a.awsS3Client.IsFileExists(tokenFileName); err != nil {
		return err
	}

	tokenPath := path.Join(sshPath, "config.yaml")
	if err := a.awsS3Client.DownloadFile(tokenPath, tokenFileName); err != nil {
		return err
	}

	b, err := os.ReadFile(tokenPath)
	if err != nil {
		return err
	}

	kc := TokenAndKubeconfig{}

	if err := yaml.Unmarshal(b, &kc); err != nil {
		return err
	}

	// setup ssh

	if err := a.SetupSSH(); err != nil {
		return err
	}
	defer a.saveForSure()

	ip, err := utils.GetOutput(path.Join(utils.Workdir, a.node.NodeId), "node-ip")
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
			"checking if node is ready"); e == nil {
			break
		}

		count++
		if count > 24 {
			return fmt.Errorf("node is not ready even after 6 minutes")
		}
		time.Sleep(time.Second * 5)
	}

	// attach to cluster as master
	cmd := fmt.Sprintf(
		"ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s/access root@%s sudo sh /tmp/k3s-install.sh server --server https://%s:6443 --token %s  --node-external-ip %s --flannel-backend wireguard-native --flannel-external-ip --disable traefik --node-name=%s",
		sshPath,
		string(ip),
		kc.ServerIp,
		strings.TrimSpace(string(kc.Token)),
		string(ip),
		a.node.NodeId,
	)

	if err := utils.ExecCmd(cmd, "attaching to cluster as a master"); err != nil {
		return err
	}

	return nil
}
