package infraclient

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

type awsProvider struct {
	accessKey    string
	accessSecret string

	serverUrl   string
	accountId   string
	sshKeyPath  string
	joinToken   string
	providerDir string
	storePath   string
	tfTemplates string
	labels      map[string]string
	taints      []string
}

type AWSNode struct {
	NodeId       string
	Region       string
	InstanceType string
	VPC          string
	AMI          string
}

type awsProviderClient interface {
	NewNode(node AWSNode) error
	DeleteNode(node AWSNode) error

	AttachNode(node AWSNode) error
	UnattachNode(node AWSNode) error

	// mkdir(folder string) error
	// rmdir(folder string) error
	// getFolder(region, nodeId string) string
	// initTFdir(region, nodeId string) error
	// applyTF(region, nodeId string, values map[string]string) error
	// destroyNode(folder string) error
	// execCmd(cmd string) error
}

// getFolder implements doProviderClient
func (a *awsProvider) getFolder(region string, nodeId string) string {
	// eg -> /path/do/blr1/acc_id/node_id
	return path.Join(a.storePath, a.providerDir, region, a.accountId, nodeId)
}

// initTFdir implements doProviderClient
func (d *awsProvider) initTFdir(node AWSNode) error {

	folder := d.getFolder(node.Region, node.NodeId)

	if err := execCmd(fmt.Sprintf("cp -r %s %s", fmt.Sprintf("%s/%s", d.tfTemplates, d.providerDir), folder), "initialize terraform"); err != nil {
		return err
	}

	cmd := exec.Command("terraform", "init", "-no-color")
	cmd.Dir = path.Join(folder, d.providerDir)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// AttachNode implements awsProviderClient
func (a *awsProvider) AttachNode(node AWSNode) error {

	out, err := getOutput(path.Join(a.getFolder(node.Region, node.NodeId), a.providerDir), "node-ip")

	if err != nil {
		return err
	}

	// wait for start
	itrationCount := 0
	for {
		e := execCmd(
			fmt.Sprintf("ssh -oStrictHostKeyChecking=no -i %s/id_rsa root@%s ls", a.sshKeyPath, out),
			"checking node ready to attach or not",
		)

		if e == nil {
			fmt.Println("[#] node is ready to attach")
			break
		}

		if itrationCount >= 100 {
			return errors.New("node is stil ready to attach after 10 minutes")
		}

		itrationCount++
		time.Sleep(time.Second * 6)
	}

	labels := func() string {
		l := ""
		for k, v := range a.labels {
			l += fmt.Sprintf(" --node-label=%s=%s", k, v)
		}
		return l
	}()

	taints := func() string {
		t := ""

		for _, v := range a.taints {
			t += fmt.Sprintf(" --node-taint %s", v)
		}

		return t
	}()

	if err = execCmd(
		fmt.Sprintf("ssh -oStrictHostKeyChecking=no -i %s/id_rsa root@%s %q",
			a.sshKeyPath,
			out,

			fmt.Sprintf(
				"curl -sfL https://get.k3s.io | sh -s - agent  --token=%s --server %s --node-ip=%s %s %s --node-name %s",
				a.joinToken,
				a.serverUrl,
				out, labels, taints, node.NodeId),
		),
		"attaching node to cluster",
	); err != nil {
		return err
	}

	return execCmd(
		fmt.Sprintf("ssh -oStrictHostKeyChecking=no -i %s/id_rsa root@%s %q",
			a.sshKeyPath,
			out,
			"history -c",
		),
		"cleaning up setup",
	)
}

// DeleteNode implements awsProviderClient
func (a *awsProvider) DeleteNode(node AWSNode) error {

	// time.Sleep(time.Minute * 2)
	values := map[string]string{}

	//TODO: remove node from cluster after drain proceed following

	values["access_key"] = a.accessKey
	values["secret_key"] = a.accessSecret

	values["ami"] = node.AMI
	values["keys-path"] = a.sshKeyPath
	values["region"] = node.Region
	values["node_id"] = node.NodeId
	values["instance_type"] = node.InstanceType

	nodetfpath := path.Join(a.getFolder(node.Region, node.NodeId), a.providerDir)

	// check if dir present
	if _, err := os.Stat(path.Join(nodetfpath, "init.sh")); err != nil && os.IsNotExist(err) {
		fmt.Println("tf state not present nothing to do")
		return nil
	}

	// get node name
	var out string
	var err error
	if out, err = getOutput(nodetfpath, "node-name"); err != nil {
		return err
	} else if strings.TrimSpace(out) == "" {
		fmt.Println("something went wrong, can't find node_name")
		return nil
	}

	if err = a.UnattachNode(node); err != nil {
		return err
	}

	// destroy node
	return destroyNode(nodetfpath, values)

}

// NewNode implements awsProviderClient
func (a *awsProvider) NewNode(node AWSNode) error {

	values := map[string]string{}

	values["access_key"] = a.accessKey
	values["secret_key"] = a.accessSecret
	values["keys-path"] = a.sshKeyPath

	values["ami"] = node.AMI
	values["region"] = node.Region
	values["node_id"] = node.NodeId
	values["instance_type"] = node.InstanceType

	// values["keys-path"] = a.sshKeyPath
	// values["accountId"] = a.accountId

	// values["do-image-id"] = node.ImageId
	// values["nodeId"] = node.NodeId
	// values["size"] = node.Size

	// making dir
	if err := mkdir(a.getFolder(node.Region, node.NodeId)); err != nil {
		return err
	}

	// initialize directory
	if err := a.initTFdir(node); err != nil {
		return err
	}

	// apply terraform
	return applyTF(path.Join(a.getFolder(node.Region, node.NodeId), a.providerDir), values)

}

// UnattachNode implements awsProviderClient
func (a *awsProvider) UnattachNode(node AWSNode) error {

	var out string
	var err error

	if out, err = getOutput(path.Join(a.getFolder(node.Region, node.NodeId), a.providerDir), "node-name"); err != nil {
		return err
	} else if strings.TrimSpace(out) == "" {
		fmt.Println("something went wrong, can't find node_name")
		return nil
	}

	if err = execCmd(fmt.Sprintf("kubectl get node %s", out), "checknode present"); err != nil {
		fmt.Println("node not found may be already deleted")
		return nil
	}

	// drain node
	if err = execCmd(fmt.Sprintf("kubectl drain %s --force --ignore-daemonsets --delete-local-data", out), "drain node to delete"); err != nil {
		return err
	}

	fmt.Println("[#] waiting 2 minutes after drain")

	// delete node
	return execCmd(fmt.Sprintf("kubectl delete node %s", out),
		"delete node from cluster")
}

type AWSProvider struct {
	AccessKey    string
	AccessSecret string
	AccountId    string
}

type AWSProviderEnv struct {
	ServerUrl  string
	SshKeyPath string

	StorePath   string
	TfTemplates string

	JoinToken string
	Labels    map[string]string
	Taints    []string
}

func NewAWSProvider(provider AWSProvider, p AWSProviderEnv) awsProviderClient {
	return &awsProvider{
		accessKey:    provider.AccessKey,
		accessSecret: provider.AccessSecret,
		accountId:    provider.AccountId,

		providerDir: "aws",
		serverUrl:   p.ServerUrl,
		joinToken:   p.JoinToken,
		sshKeyPath:  p.SshKeyPath,
		storePath:   p.StorePath,
		tfTemplates: p.TfTemplates,
		labels:      p.Labels,
		taints:      p.Taints,
	}
}
