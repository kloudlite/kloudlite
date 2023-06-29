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

type doProviderClient interface {
	NewNode(node DoNode) error
	DeleteNode(node DoNode) error

	AttachNode(node DoNode) error
	UnattachNode(node DoNode) error

	// mkdir(folder string) error
	// rmdir(folder string) error
	// getFolder(region, nodeId string) string
	// initTFdir(region, nodeId string) error
	// applyTF(region, nodeId string, values map[string]string) error
	// destroyNode(folder string) error
	// execCmd(cmd string) error
}

type DoNode struct {
	Region  string
	Size    string
	NodeId  string
	ImageId string
}

type joinTokenSecret struct {
	JoinToken   string `json:"joinToken" yaml:"joinToken"`
	EndpointUrl string `json:"endpointUrl" yaml:"endPointUrl"`
}

type doProvider struct {
	apiToken    string
	accountId   string
	providerDir string
	storePath   string
	tfTemplates string
	SSHPath     string
	PubKey      string
	labels      map[string]string
	taints      []string
	secrets     string
}

// getFolder implements doProviderClient
func (d *doProvider) getFolder(region string, nodeId string) string {
	// eg -> /path/do/blr1/acc_id/node_id
	return path.Join(d.storePath, d.accountId, d.providerDir, region, nodeId)
}

// initTFdir implements doProviderClient
func (d *doProvider) initTFdir(node DoNode) error {

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

// NewNode implements doProviderClient
func (d *doProvider) NewNode(node DoNode) error {
	values := map[string]string{}

	values["cluster-id"] = CLUSTER_ID

	// values["keys-path"] = d.sshKeyPath
	values["do-token"] = d.apiToken
	values["accountId"] = d.accountId

	values["do-image-id"] = "ubuntu-22-10-x64"
	values["nodeId"] = node.NodeId
	values["size"] = node.Size
	values["keys-path"] = d.SSHPath

	// making dir
	if err := mkdir(d.getFolder(node.Region, node.NodeId)); err != nil {
		return err
	}

	// initialize directory
	if err := d.initTFdir(node); err != nil {
		return err
	}

	tfPath := path.Join(d.getFolder(node.Region, node.NodeId), d.providerDir)

	// apply terraform
	return applyTF(tfPath, values)
}

// DeleteNode implements ProviderClient
func (d *doProvider) DeleteNode(node DoNode) error {
	// time.Sleep(time.Minute * 2)
	values := map[string]string{}

	//TODO: remove node from cluster after drain proceed following

	values["cluster-id"] = CLUSTER_ID

	// values["keys-path"] = d.sshKeyPath
	values["do-token"] = d.apiToken
	values["accountId"] = d.accountId

	// values["do-image-id"] = node.ImageId
	values["do-image-id"] = "ubuntu-22-10-x64"
	values["nodeId"] = node.NodeId
	values["keys-path"] = d.SSHPath

	nodetfpath := path.Join(d.getFolder(node.Region, node.NodeId), d.providerDir)

	// check if dir present
	if _, err := os.Stat(path.Join(nodetfpath, "init.sh")); err != nil && os.IsNotExist(err) {
		fmt.Println("tf state not present nothing to do")
		return nil
	}

	// get node name
	var out []byte
	var err error
	if out, err = getOutput(nodetfpath, "node-name"); err != nil {
		return err
	} else if strings.TrimSpace(string(out)) == "" {
		fmt.Println("something went wrong, can't find node_name")
		return nil
	}

	// destroy node
	return destroyNode(nodetfpath, values)
}

// AttachNode implements ProviderClient
func (d *doProvider) AttachNode(node DoNode) error {

	var out, secretYaml []byte
	var err error

	if out, err = getOutput(path.Join(d.getFolder(node.Region, node.NodeId), d.providerDir), "node-ip"); err != nil {
		return err
	}

	if secretYaml, err = base64.StdEncoding.DecodeString(d.secrets); err != nil {
		fmt.Println("here", d.secrets)
		return err
	}

	var sec joinTokenSecret

	if err = yaml.Unmarshal(secretYaml, &sec); err != nil {
		return err
	}

	labels := func() []string {
		l := []string{}
		for k, v := range d.labels {
			l = append(l, fmt.Sprintf("--node-label %s=%s", k, v))
		}
		l = append(l, fmt.Sprintf("--node-label %s=%s", "kloudlite.io/public-ip", string(out)))
		return l
	}()

	// fmt.Println(labels)

	// "hostname": node.NodeId,
	// "labels": strings.Join(labels, ","),
	//check is node ready
	count := 0

	for {
		if e := execCmd(
			fmt.Sprintf("ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s root@%s ls",
				fmt.Sprintf("%v/access", d.SSHPath),
				string(out)),
			"checking if node is ready"); e == nil {
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
		fmt.Sprintf("ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s root@%s sudo sh /tmp/k3s-install.sh agent --server %s --token %s %s --node-name %s --node-external-ip %s",
			fmt.Sprintf("%v/access", d.SSHPath), string(out), sec.EndpointUrl, sec.JoinToken,
			strings.Join(labels, " "), node.NodeId, string(out)),
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

	return nil
}

// UnattachNode implements doProviderClient
func (d *doProvider) UnattachNode(node DoNode) error {
	var out []byte
	var err error

	if out, err = getOutput(path.Join(d.getFolder(node.Region, node.NodeId), d.providerDir), "node-name"); err != nil {
		return err
	} else if strings.TrimSpace(string(out)) == "" {
		fmt.Println("something went wrong, can't find node_name")
		return nil
	}

	if err = execCmd(fmt.Sprintf("kubectl get node %s", out), "checking if node attached"); err != nil {
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

type DoProvider struct {
	ApiToken  string
	AccountId string
}

type DoProviderEnv struct {
	StorePath   string
	TfTemplates string

	SSHPath string
	Secrets string
	Labels  map[string]string
	Taints  []string
}

func NewDOProvider(provider DoProvider, p DoProviderEnv) doProviderClient {
	return &doProvider{
		secrets:     p.Secrets,
		apiToken:    provider.ApiToken,
		accountId:   provider.AccountId,
		providerDir: "do",
		storePath:   p.StorePath,
		tfTemplates: p.TfTemplates,
		labels:      p.Labels,
		taints:      p.Taints,
		SSHPath:     p.SSHPath,
	}
}
