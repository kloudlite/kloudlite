package infraclient

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"kloudlite.io/pkg/infraClient/templates"
)

type doProviderClient interface {
	NewNode(node DoNode) error
	DeleteNode(node DoNode) error

	AttachNode(node DoNode) error
	// UnattachNode(node DoNode) error

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

type talosSecret struct {
	Secrets struct {
		Endpoint   string `json:"endpoint" yaml:"endpoint"`
		EndpointIp string `json:"endpointIp" yaml:"endpointIp"`
		TConfig    struct {
			Ca   string `json:"ca" yaml:"ca"`
			Cert string `json:"cert" yaml:"cert"`
			Key  string `json:"key" yaml:"key"`
		} `json:"talosconfig" yaml:"talosconfig"`
		Machine struct {
			Token       string `json:"token" yaml:"token"`
			Type        string `json:"type" yaml:"type"`
			Certifacate string `json:"cert" yaml:"cert"`
			Key         string `json:"key" yaml:"key"`
		} `json:"machine" yaml:"machine"`

		Cluster struct {
			Name             string `json:"name" yaml:"name"`
			Id               string `json:"id" yaml:"id"`
			Secret           string `json:"secret" yaml:"secret"`
			Token            string `json:"token" yaml:"token"`
			EncryptionSecret string `json:"encryptionSecret" yaml:"encryptionSecret"`
			Certifacate      string `json:"cert" yaml:"cert"`
			Key              string `json:"key" yaml:"key"`

			Aggregator struct {
				Certifacate string `json:"cert" yaml:"cert"`
				Key         string `json:"key" yaml:"key"`
			} `json:"aggregator" yaml:"aggregator"`
			ServiceAccountKey string `json:"serviceAccountKey" yaml:"serviceAccountKey"`
		} `json:"cluster" yaml:"cluster"`
		Etcd struct {
			Certifacate string `json:"cert" yaml:"cert"`
			Key         string `json:"key" yaml:"key"`
		} `json:"etcd" yaml:"etcd"`
	} `json:"secrets" yaml:"secrets"`
}

type doProvider struct {
	apiToken    string
	accountId   string
	providerDir string
	storePath   string
	tfTemplates string
	labels      map[string]string
	taints      []string
	secrets     string
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

	cmd := exec.Command("terraform", "init", "-no-color")
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

	values["do-image-id"] = node.ImageId
	values["nodeId"] = node.NodeId

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

	if err := d.UnattachNode(node); err != nil {
		return err
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
		return err
	}

	var sec talosSecret

	if err = yaml.Unmarshal(secretYaml, &sec); err != nil {
		return err
	}

	var config, tConf []byte

	labels := func() []string {
		l := []string{}
		for k, v := range d.labels {
			l = append(l, fmt.Sprintf("%s=%s", k, v))
		}
		l = append(l, fmt.Sprintf("%s=%s", "kloudlite.io/public-ip", string(out)))
		return l
	}()

	sec.Secrets.Machine.Type = "controlplane"

	switch sec.Secrets.Machine.Type {
	case "worker":
		if config, err = templates.Parse(templates.WorkerConfig, map[string]interface{}{
			"hostname": node.NodeId,

			"machine-cert":  sec.Secrets.Machine.Certifacate,
			"machine-type":  sec.Secrets.Machine.Type,
			"machine-token": sec.Secrets.Machine.Token,

			"endpoint": sec.Secrets.Endpoint,

			"cluster-id":     sec.Secrets.Cluster.Id,
			"cluster-token":  sec.Secrets.Cluster.Token,
			"cluster-cert":   sec.Secrets.Cluster.Certifacate,
			"cluster-secret": sec.Secrets.Cluster.Secret,
			"labels":         strings.Join(labels, ","),
		}); err != nil {
			return err
		}
	case "controlplane":
		if config, err = templates.Parse(templates.ControlePlaneConfig, map[string]interface{}{
			"hostname": node.NodeId,

			"machine-cert":  sec.Secrets.Machine.Certifacate,
			"machine-key":   sec.Secrets.Machine.Key,
			"machine-token": sec.Secrets.Machine.Token,
			"machine-type":  sec.Secrets.Machine.Type,

			"endpoint":    sec.Secrets.Endpoint,
			"endpoint-ip": sec.Secrets.EndpointIp,

			"cluster-name":                sec.Secrets.Cluster.Name,
			"cluster-token":               sec.Secrets.Cluster.Token,
			"cluster-encryption-secret":   sec.Secrets.Cluster.EncryptionSecret,
			"cluster-id":                  sec.Secrets.Cluster.Id,
			"cluster-secret":              sec.Secrets.Cluster.Secret,
			"cluster-key":                 sec.Secrets.Cluster.Key,
			"cluster-cert":                sec.Secrets.Cluster.Certifacate,
			"cluster-aggregator-cert":     sec.Secrets.Cluster.Aggregator.Certifacate,
			"cluster-aggregator-key":      sec.Secrets.Cluster.Aggregator.Key,
			"cluster-service-account-key": sec.Secrets.Cluster.ServiceAccountKey,

			"etcd-key":  sec.Secrets.Etcd.Key,
			"etcd-cert": sec.Secrets.Etcd.Certifacate,

			"labels": strings.Join(labels, ","),
		}); err != nil {
			return err
		}
	}

	talosConfigP := path.Join(d.getFolder(node.Region, node.NodeId), "talosconfig.yml")
	if err = ioutil.WriteFile(talosConfigP, config, fs.ModePerm); err != nil {
		return err
	}

	if tConf, err = templates.Parse(templates.TalosConfig, map[string]interface{}{
		"endpoint":     string(out),
		"cluster-name": sec.Secrets.Cluster.Name,
		"ca":           sec.Secrets.TConfig.Ca,
		"cert":         sec.Secrets.TConfig.Cert,
		"key":          sec.Secrets.TConfig.Key,
	}); err != nil {
		return err
	} else {
		if err = ioutil.WriteFile(talosConfigP, tConf, fs.ModePerm); err != nil {
			return err
		}
	}

	p := path.Join(d.getFolder(node.Region, node.NodeId), "config.yml")
	if err = ioutil.WriteFile(p, config, fs.ModePerm); err != nil {
		return err
	}

	count := 1
	for {

		time.Sleep(time.Second * 6)
		if err = execCmd(fmt.Sprintf("talosctl apply-config --insecure --nodes %s --file %s", string(out), p), "attaching node to cluster"); err != nil {

			if err = execCmd(fmt.Sprintf("talosctl stats -n %s --talosconfig %s", string(out), talosConfigP), "checking is node ready"); err == nil {
				return nil
			}

			count++
			continue
		}

		if count >= 10 {
			return errors.New("faild to apply config after 10 attempts")
		}

		return nil
	}

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
	}
}
