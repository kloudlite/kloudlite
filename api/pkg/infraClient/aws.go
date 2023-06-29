package infraclient

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type awsProvider struct {
	accessKey    string
	accessSecret string

	SSHPath     string
	accountId   string
	secrets     string
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
	ImageId      string
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
	// eg -> /path/acc_id/do/blr1/node_id/do

	return path.Join(a.storePath, a.accountId, a.providerDir, region, nodeId)
}

// initTFdir implements doProviderClient
func (d *awsProvider) initTFdir(node AWSNode) error {

	folder := d.getFolder(node.Region, node.NodeId)

	if err := execCmd(fmt.Sprintf("cp -r %s %s", fmt.Sprintf("%s/%s", d.tfTemplates, d.providerDir), folder), "initialize terraform"); err != nil {
		return err
	}

	cmd := exec.Command("terraform", "init")
	cmd.Dir = path.Join(folder, d.providerDir)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

type TalosAmi struct {
	Cloud   string
	Version string
	Region  string
	Arch    string
	Type    string
	Id      string
}

// NewNode implements awsProviderClient
func (a *awsProvider) NewNode(node AWSNode) error {

	values := map[string]string{}

	values["access_key"] = a.accessKey
	values["secret_key"] = a.accessSecret

	values["region"] = node.Region
	values["node_id"] = node.NodeId
	values["instance_type"] = node.InstanceType
	values["keys-path"] = a.SSHPath
	values["ami"] = node.ImageId

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

// AttachNode implements awsProviderClient
func (a *awsProvider) AttachNode(node AWSNode) error {

	var out, secretYaml []byte
	var err error

	if out, err = getOutput(path.Join(a.getFolder(node.Region, node.NodeId), a.providerDir), "node-ip"); err != nil {
		return err
	}

	if secretYaml, err = base64.StdEncoding.DecodeString(a.secrets); err != nil {
		return err
	}

	var sec joinTokenSecret

	if err = yaml.Unmarshal(secretYaml, &sec); err != nil {
		return err
	}

	labels := func() []string {
		l := []string{}
		for k, v := range a.labels {
			l = append(l, fmt.Sprintf("--node-label %s=%s", k, v))
		}
		l = append(l, fmt.Sprintf("--node-label %s=%s", "kloudlite.io/public-ip", string(out)))
		return l
	}()

	count := 0

	for {
		if e := execCmd(
			fmt.Sprintf("ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s root@%s ls",
				fmt.Sprintf("%v/access", a.SSHPath),
				string(out)),
			"checking if node is ready "); e == nil {
			break
		}

		count++
		if count > 24 {
			return fmt.Errorf("node is not ready even after 6 minutes")
		}
		time.Sleep(time.Second * 15)
	}

	if err = execCmd(fmt.Sprintf("kubectl get node %s", node.NodeId), "checking if node attached"); err == nil {
		fmt.Println("node already attached. clean exit")
		return nil
	}

	// // install k3s
	// if e := execCmd(
	// 	fmt.Sprintf("ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s root@%s sudo sh /tmp/k3s-install.sh",
	// 		fmt.Sprintf("%v/access", d.SSHPath), string(out)),
	// 	""); e != nil {
	// 	return e
	// }

	// attach node
	if e := execCmd(
		fmt.Sprintf("ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s root@%s sudo sh /tmp/k3s-install.sh agent --server %s --token %s %s --node-name %s --node-external-ip %s --node-ip %s",
			fmt.Sprintf("%v/access", a.SSHPath), string(out), sec.EndpointUrl, sec.JoinToken,
			strings.Join(labels, " "), node.NodeId, string(out), string(out)),
		"attaching to cluster"); e != nil {
		return e
	}

	count = 0
	for {
		if err = execCmd(fmt.Sprintf("kubectl get node %s", node.NodeId), "checking if node attached"); err == nil {
			fmt.Println("node attached successfully.")
			break
		}

		count++
		if count > 24 {
			return fmt.Errorf("node not attached even after 6minutes")
		}
		time.Sleep(time.Second * 15)
	}

	// "hostname": node.NodeId,
	// "labels":         strings.Join(labels, ","),
	// TODO: needs to AttachNode here
	return nil

}

// DeleteNode implements awsProviderClient
func (a *awsProvider) DeleteNode(node AWSNode) error {
	var err error

	// time.Sleep(time.Minute * 2)
	values := map[string]string{}

	//TODO: remove node from cluster after drain proceed following

	values["access_key"] = a.accessKey
	values["secret_key"] = a.accessSecret

	values["region"] = node.Region
	values["node_id"] = node.NodeId
	values["instance_type"] = node.InstanceType
	values["keys-path"] = a.SSHPath
	values["ami"] = node.ImageId

	nodetfpath := path.Join(a.getFolder(node.Region, node.NodeId), a.providerDir)

	// check if dir present
	if _, e := os.Stat(path.Join(nodetfpath, "init.sh")); e != nil && os.IsNotExist(e) {
		fmt.Println("tf state not present nothing to do")
		return nil
	}

	// get node name
	var out []byte
	if out, err = getOutput(nodetfpath, "node-name"); err != nil {
		return err
	} else if strings.TrimSpace(string(out)) == "" {
		fmt.Println("something went wrong, can't find node_name")
		return nil
	}

	// destroy node
	return destroyNode(nodetfpath, values)

}

func (a *awsProvider) UnattachNode(node AWSNode) error {
	var out []byte
	var err error

	if out, err = getOutput(path.Join(a.getFolder(node.Region, node.NodeId), a.providerDir), "node-name"); err != nil {
		return err
	} else if strings.TrimSpace(string(out)) == "" {
		fmt.Println("something went wrong, can't find node_name")
		return nil
	}

	if err = execCmd(fmt.Sprintf("kubectl get node %s", out), "checknode present"); err != nil {
		fmt.Println("node not found may be already deleted")
		return nil
	}

	// drain node
	if err = execCmd(fmt.Sprintf("kubectl taint nodes %s force=delete:NoExecute", node.NodeId), "drain node to delete"); err != nil {
		return err
	}

	fmt.Println("[#] waiting 10 seconds after drain")
	time.Sleep(time.Second * 10)

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
	StorePath   string
	TfTemplates string

	Labels  map[string]string
	Taints  []string
	Secrets string
	SSHPath string
}

func NewAWSProvider(provider AWSProvider, p AWSProviderEnv) awsProviderClient {
	return &awsProvider{
		accessKey:    provider.AccessKey,
		accessSecret: provider.AccessSecret,
		accountId:    provider.AccountId,

		providerDir: "aws",
		secrets:     p.Secrets,
		storePath:   p.StorePath,
		tfTemplates: p.TfTemplates,
		labels:      p.Labels,
		taints:      p.Taints,
		SSHPath:     p.SSHPath,
	}
}
