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

func (a AwsClient) AddWorker(ctx context.Context) error {
	// fetch token

	sshPath := path.Join("/tmp/ssh", a.accountName)

	if err := a.ensurePaths(); err != nil {
		return err
	}

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

	if a.node.ProvisionMode == "spot" {
		if err := a.writeNodeConfig(kc); err != nil {
			return err
		}
	}

	// setup ssh

	if err := a.SetupSSH(); err != nil {
		return err
	}
	defer a.saveForSure()

	// create node and wait for ready
	if err := a.NewNode(ctx); err != nil {
		return err
	}

	if a.node.ProvisionMode == "spot" {
		return nil
	}

	ip, err := utils.GetOutput(path.Join(utils.Workdir, *a.node.NodeName), "node-ip")
	if err != nil {
		return err
	}

	count := 0

	for {
		if e := utils.ExecCmd(
			fmt.Sprintf("ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s root@%s ls",
				fmt.Sprintf("%s/access", sshPath),
				string(ip),
			),
			"checking if node ready"); e == nil {
			break
		}

		count++
		if count > 24 {
			return fmt.Errorf("node is not ready even after 6 minutes")
		}
		time.Sleep(time.Second * 5)
	}

	labels := func() []string {
		l := []string{}
		for k, v := range map[string]string{
			"kloudlite.io/public-ip": string(ip),
		} {
			l = append(l, fmt.Sprintf("--node-label %s=%s", k, v))
		}

		for k, v := range a.labels {
			l = append(l, fmt.Sprintf("--node-label %s=%s", k, v))
		}
		return l
	}()

	// attach to cluster as workernode

	cmd := fmt.Sprintf(
		"ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s/access root@%s sudo sh /tmp/k3s-install.sh agent --server https://%s:6443 --token=%s --node-external-ip %s --node-name %s %s %s",
		sshPath,
		ip,
		kc.ServerIp,
		strings.TrimSpace(string(kc.Token)),
		ip,
		*a.node.NodeName,
		strings.Join(labels, " "),
		func() string {
			if a.node.IsGpu != nil && *a.node.IsGpu {
				// return "--docker"
				// return "--docker"
				return ""
			}
			return ""
		}(),
	)

	if err := utils.ExecCmd(cmd, "attaching to cluster as a worker node"); err != nil {
		return err
	}

	return nil
}
