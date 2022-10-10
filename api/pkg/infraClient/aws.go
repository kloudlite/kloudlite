package infraclient

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"kloudlite.io/pkg/infraClient/templates"
)

type awsProvider struct {
	accessKey    string
	accessSecret string

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
}

type awsProviderClient interface {
	NewNode(node AWSNode) error
	DeleteNode(node AWSNode) error

	AttachNode(node AWSNode) error
	// UnattachNode(node AWSNode) error

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

type TalosAmi struct {
	Cloud   string
	Version string
	Region  string
	Arch    string
	Type    string
	Id      string
}

func getAmi(region string) (string, error) {

	req, err := http.NewRequest("GET", "https://github.com/siderolabs/talos/releases/download/v1.2.3/cloud-images.json", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	amis := []TalosAmi{}

	if err := json.Unmarshal(b, &amis); err != nil {
		return "", err
	}

	ami := ""
	for _, ta := range amis {
		if ta.Region == region && ta.Arch == "amd64" {
			ami = ta.Id
		}
	}
	if ami == "" {
		return "", errors.New("can't find ami for talos")
	}

	return ami, nil
}

// NewNode implements awsProviderClient
func (a *awsProvider) NewNode(node AWSNode) error {

	ami, err := getAmi(node.Region)
	if err != nil {
		return err
	}

	values := map[string]string{}

	values["access_key"] = a.accessKey
	values["secret_key"] = a.accessSecret

	values["ami"] = ami
	values["region"] = node.Region
	values["node_id"] = node.NodeId
	values["instance_type"] = node.InstanceType

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

	var sec talosSecret

	if err = yaml.Unmarshal(secretYaml, &sec); err != nil {
		return err
	}

	var config, tConf []byte

	labels := func() []string {
		l := []string{}
		for k, v := range a.labels {
			l = append(l, fmt.Sprintf("%s=%s", k, v))
		}
		l = append(l, fmt.Sprintf("%s=%s", "kloudlite.io/public-ip", string(out)))
		return l
	}()

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

	talosConfigP := path.Join(a.getFolder(node.Region, node.NodeId), "talosconfig.yml")
	if err = ioutil.WriteFile(talosConfigP, config, fs.ModePerm); err != nil {
		return err
	}

	if tConf, err = templates.Parse(templates.TalosConfig, map[string]interface{}{
		"endpoint":    string(out),
		"clusterName": sec.Secrets.Cluster.Name,
		"ca":          sec.Secrets.TConfig.Ca,
		"cert":        sec.Secrets.TConfig.Cert,
		"key":         sec.Secrets.TConfig.Key,
	}); err != nil {
		return err
	} else {
		if err = ioutil.WriteFile(talosConfigP, tConf, fs.ModePerm); err != nil {
			return err
		}
	}

	p := path.Join(a.getFolder(node.Region, node.NodeId), "config.yml")
	if err = ioutil.WriteFile(p, config, fs.ModePerm); err != nil {
		return err
	}

	count := 1
	for {

		time.Sleep(time.Second * 6)
		if err = execCmd(fmt.Sprintf("talosctl apply-config --insecure --nodes %s --file %s", string(out), p), ""); err != nil {

			if err = execCmd(fmt.Sprintf("talosctl stats -n %s --talosconfig %s", string(out), talosConfigP), ""); err == nil {
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

// DeleteNode implements awsProviderClient
func (a *awsProvider) DeleteNode(node AWSNode) error {
	var err error

	ami, err := getAmi(node.Region)
	if err != nil {
		return err
	}

	// time.Sleep(time.Minute * 2)
	values := map[string]string{}

	//TODO: remove node from cluster after drain proceed following

	values["access_key"] = a.accessKey
	values["secret_key"] = a.accessSecret

	values["ami"] = ami
	values["region"] = node.Region
	values["node_id"] = node.NodeId
	values["instance_type"] = node.InstanceType

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

	if err := a.UnattachNode(node); err != nil {
		return err
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
	}
}
