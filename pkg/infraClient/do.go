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

type doProvider struct {
	serverUrl   string
	apiToken    string
	accountId   string
	sshKeyPath  string
	joinToken   string
	providerDir string
	storePath   string
	tfTemplates string
	labels      map[string]string
}

// getFolder implements doProviderClient
func (d *doProvider) getFolder(region string, nodeId string) string {
	// eg -> /path/do/blr1/acc_id/node_id
	return path.Join(d.storePath, d.providerDir, region, d.accountId, nodeId)
}

// initTFdir implements doProviderClient
func (d *doProvider) initTFdir(node DoNode) error {

	folder := d.getFolder(node.Region, node.NodeId)

	if err := execCmd(fmt.Sprintf("cp -r %s %s", fmt.Sprintf("%s/%s", d.tfTemplates, d.providerDir), folder), "initialize terraform"); err != nil {
		return err
	}

	cmd := exec.Command("terraform", "init")
	cmd.Dir = path.Join(folder, "do")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// NewNode implements doProviderClient
func (d *doProvider) NewNode(node DoNode) error {
	values := map[string]string{}

	values["cluster-id"] = CLUSTER_ID

	values["keys-path"] = d.sshKeyPath
	values["do-token"] = d.apiToken
	values["accountId"] = d.accountId

	values["do-image-id"] = node.ImageId
	values["nodeId"] = node.NodeId
	values["size"] = node.Size

	// making dir
	if err := mkdir(d.getFolder(node.Region, node.NodeId)); err != nil {
		return err
	}

	// initialize directory
	if err := d.initTFdir(node); err != nil {
		return err
	}

	// apply terraform
	return applyTF(path.Join(d.getFolder(node.Region, node.NodeId), "do"), values)
}

// UnattachNode implements doProviderClient
func (d *doProvider) UnattachNode(node DoNode) error {
	var out string
	var err error

	if out, err = getOutput(path.Join(d.getFolder(node.Region, node.NodeId), "do"), "node-name"); err != nil {
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
	if err = execCmd(fmt.Sprintf("kubectl drain %s --ignore-daemonsets --delete-local-data", out), "drain node to delete"); err != nil {
		return err
	}

	fmt.Println("[#] waiting 2 minutes after drain")

	// delete node
	return execCmd(fmt.Sprintf("kubectl delete node %s", out),
		"delete node from cluster")
}

// DeleteNode implements ProviderClient
func (d *doProvider) DeleteNode(node DoNode) error {
	// time.Sleep(time.Minute * 2)
	values := map[string]string{}

	//TODO: remove node from cluster after drain proceed following

	values["cluster-id"] = CLUSTER_ID

	values["keys-path"] = d.sshKeyPath
	values["do-token"] = d.apiToken
	values["accountId"] = d.accountId

	values["do-image-id"] = node.ImageId
	values["nodeId"] = node.NodeId

	nodetfpath := path.Join(d.getFolder(node.Region, node.NodeId), "do")

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

	if err = d.UnattachNode(node); err != nil {
		return err
	}

	// destroy node
	return destroyNode(nodetfpath, values)

	// remove node tf state
	// return rmdir(d.getFolder(node.Region, node.NodeId))
}

// AttachNode implements ProviderClient
func (d *doProvider) AttachNode(node DoNode) error {

	out, err := getOutput(path.Join(d.getFolder(node.Region, node.NodeId), "do"), "node-ip")

	if err != nil {
		return err
	}

	// wait for start
	itrationCount := 0
	for {
		e := execCmd(
			fmt.Sprintf("ssh -oStrictHostKeyChecking=no -i %s/id_rsa root@%s ls", d.sshKeyPath, out),
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
		for k, v := range d.labels {
			l += fmt.Sprintf(" --node-label=%s=%s", k, v)
		}
		return l
	}()

	if err = execCmd(
		fmt.Sprintf("ssh -oStrictHostKeyChecking=no -i %s/id_rsa root@%s %q",
			d.sshKeyPath,
			out,

			fmt.Sprintf(
				"curl -sfL https://get.k3s.io | sh -s - agent  --token=%s --server %s --node-ip=%s %s",
				d.joinToken,
				d.serverUrl,
				out, labels),
		),
		"attaching node to cluster",
	); err != nil {
		return err
	}

	return execCmd(
		fmt.Sprintf("ssh -oStrictHostKeyChecking=no -i %s/id_rsa root@%s %q",
			d.sshKeyPath,
			out,
			"history -c",
		),
		"cleaning up setup",
	)

}

type DoProvider struct {
	ApiToken  string
	AccountId string
}

type ProviderEnv struct {
	ServerUrl  string
	SshKeyPath string

	StorePath   string
	TfTemplates string

	JoinToken string
	Labels    map[string]string
}

func NewDOProvider(provider DoProvider, p ProviderEnv) doProviderClient {
	return &doProvider{
		serverUrl:   p.ServerUrl,
		joinToken:   p.JoinToken,
		apiToken:    provider.ApiToken,
		accountId:   provider.AccountId,
		sshKeyPath:  p.SshKeyPath,
		providerDir: "do",
		storePath:   p.StorePath,
		tfTemplates: p.TfTemplates,
		labels:      p.Labels,
	}
}
