package aws

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"gopkg.in/yaml.v2"

	"kloudlite.io/apps/nodectrl/internal/domain/common"
	"kloudlite.io/apps/nodectrl/internal/domain/utils"
	awss3 "kloudlite.io/pkg/aws-s3"
)

type AwsProviderConfig struct {
	AccessKey    string `yaml:"accessKey"`
	AccessSecret string `yaml:"accessSecret"`
	AccountName  string `yaml:"accountName"`
}
type AWSNodeConfig struct {
	clustersv1.AWSNodeConfig `json:",inline"`
	NodeName                 string `json:"nodeName"`
}

type AwsClient struct {
	node        AWSNodeConfig
	awsS3Client awss3.AwsS3

	accessKey    string
	accessSecret string
	accountName  string

	tfTemplates string
	labels      map[string]string
	taints      []string
}

type TokenAndKubeconfig struct {
	Token       string `json:"token"`
	Kubeconfig  string `json:"kubeconfig"`
	ServerIp    string `json:"serverIp"`
	MasterToken string `json:"masterToken"`
}

type NodeConfig struct {
	ServerIP string            `yaml:"serverIp"`
	Token    string            `yaml:"token"`
	NodeName string            `yaml:"nodeName"`
	Taints   []string          `yaml:"taints"`
	Labels   map[string]string `yaml:"labels"`
}

func (a AwsClient) ensurePaths() error {
	workDir := path.Join("/tmp", a.node.NodeName)
	if _, err := os.Stat(workDir); err != nil {
		if err := os.Mkdir(workDir, os.ModePerm); err != nil {
			return err
		}
	}

	const sshDir = "/tmp/ssh"
	sshPath := path.Join(sshDir, a.accountName)
	if _, err := os.Stat(sshDir); err != nil {
		if err := os.Mkdir(sshDir, os.ModePerm); err != nil {
			return err
		}
	}

	if _, err := os.Stat(sshPath); err != nil {
		if err := os.Mkdir(sshPath, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func (a AwsClient) writeNodeConfig(kc TokenAndKubeconfig) error {
	if err := a.ensurePaths(); err != nil {
		return err
	}

	const sshDir = "/tmp/ssh"
	sshPath := path.Join(sshDir, a.accountName)
	dataPath := path.Join(sshPath, "data.yaml")

	out, err := yaml.Marshal(NodeConfig{
		ServerIP: kc.ServerIp,
		Token:    kc.Token,
		NodeName: a.node.NodeName,
		Taints:   []string{},
		Labels:   map[string]string{},
	})
	if err != nil {
		return err
	}

	return os.WriteFile(dataPath, out, os.ModePerm)
}

func (a AwsClient) SetupSSH() error {
	const sshDir = "/tmp/ssh"
	if err := a.ensurePaths(); err != nil {
		return err
	}

	destDir := path.Join(sshDir, a.accountName)
	fileName := fmt.Sprintf("%s.zip", a.accountName)
	if err := a.awsS3Client.IsFileExists(fileName); err != nil {
		if _, err := os.Stat(destDir); err == nil {
			if err := os.RemoveAll(destDir); err != nil {
				return err
			}
		}

		if e := os.Mkdir(destDir, os.ModePerm); e != nil {
			return e
		}

		privateKeyBytes, publicKeyBytes, err := utils.GenerateKeys()
		if err != nil {
			return err
		}

		if err := os.WriteFile(fmt.Sprintf("%s/access.pub", destDir), publicKeyBytes, os.ModePerm); err != nil {
			return err
		}

		if err := os.WriteFile(fmt.Sprintf("%s/access", destDir), privateKeyBytes, 0400); err != nil {
			return err
		}

		return nil
	}

	if err := os.RemoveAll(destDir); err != nil {
		return err
	}

	err := a.awsS3Client.DownloadFile(path.Join(sshDir, fileName), fileName)
	if err != nil {
		return err
	}

	_, err = utils.Unzip(path.Join(sshDir, fileName), sshDir)
	if err != nil {
		return err
	}

	return nil
}

func (a AwsClient) saveForSure() error {
	count := 0
	for {
		if err := a.saveSSH(); err == nil {
			return nil
		}
		if count >= 10 {
			return fmt.Errorf("coudn't save the state")
		}

		time.Sleep(time.Second * 20)
		count++
	}
}

func (a AwsClient) saveSSH() error {
	const sshDir = "/tmp/ssh"
	destDir := path.Join(sshDir, a.accountName)
	fileName := fmt.Sprintf("%s.zip", a.accountName)

	if err := utils.ZipSource(destDir, path.Join(sshDir, fileName)); err != nil {
		return err
	}

	if err := a.awsS3Client.UploadFile(path.Join(sshDir, fileName), fileName); err != nil {
		return err
	}

	return nil
}

func (a AwsClient) SaveToDbGuranteed(ctx context.Context) {
	for {
		if err := utils.SaveToDb(a.node.NodeName, a.awsS3Client); err == nil {
			break
		} else {
			fmt.Println(err)
		}
		time.Sleep(time.Second * 20)
	}
}

func (a AwsClient) getAwsTemplatePath() string {
	return path.Join(a.tfTemplates, func() string {
		switch a.node.ProvisionMode {
		case "spot":
			return "aws-spot"
		default:
			return "aws"
		}
	}(),
	)
}

// NewNode implements ProviderClient
func (a AwsClient) NewNode(ctx context.Context) error {
	sshPath := path.Join("/tmp/ssh", a.accountName)
	values, err := parseValues(a, sshPath)
	if err != nil {
		return err
	}

	if err := utils.MakeTfWorkFileReady(a.node.NodeName, a.getAwsTemplatePath(), a.awsS3Client, true); err != nil {
		return err
	}

	// upload the final state to the db, upsert if db is already present
	defer a.SaveToDbGuranteed(ctx)

	// apply the tf file
	if err := func() error {
		if err := utils.InitTFdir(path.Join(utils.Workdir, a.node.NodeName)); err != nil {
			return err
		}

		if err := utils.ApplyTF(path.Join(utils.Workdir, a.node.NodeName), values); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	return nil
}

// DeleteNode implements ProviderClient
func (a AwsClient) DeleteNode(ctx context.Context, force bool) error {
	sshPath := path.Join("/tmp/ssh", a.accountName)

	// ensure all required paths created if not exists
	if err := a.ensurePaths(); err != nil {
		return err
	}

	// parse all values from the awsclient to pass as tf variables
	values, err := parseValues(a, sshPath)
	if err != nil {
		return err
	}

	// fetch kubeconfig
	kc, err := func() ([]byte, error) {
		sshPath := path.Join("/tmp/ssh", a.accountName)

		configFileName := fmt.Sprintf("%s-config.yaml", a.accountName)
		if err := a.awsS3Client.IsFileExists(configFileName); err != nil {
			return nil, err
		}

		configPath := path.Join(sshPath, "config.yaml")
		if err := a.awsS3Client.DownloadFile(configPath, configFileName); err != nil {
			return nil, err
		}

		b, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}

		kc := TokenAndKubeconfig{}
		if err := yaml.Unmarshal(b, &kc); err != nil {
			return nil, err
		}

		out, err := yaml.Marshal(kc.Kubeconfig)
		return out, err
	}()
	if err != nil {
		return err
	}

	// drain and delete node befor destroying
	if a.node.ProvisionMode != "spot" {
		// drain
		if err := utils.Drain(kc, a.node.NodeName); err != nil {
			if !force {
				fmt.Println(err.Error())
				fmt.Println(utils.ColorText("ignoring error because of force delete", 2))

				return nil
			}
			return err
		}
		time.Sleep(time.Second * 15)

		// delete from cluster
		if err := utils.DeleteNode(kc, a.node.NodeName); err != nil {
			if !force {
				fmt.Println(err.Error())
				fmt.Println(utils.ColorText("ignoring error because of force delete", 2))
				return nil

			}
			return err
		}
	}

	// setup ssh so nodes can be accesed [generate rsa for first time]
	if err := a.SetupSSH(); err != nil {
		return err
	}

	// initiate the tf_template and init the tf_directory to sync the tf_plugin
	if err := utils.MakeTfWorkFileReady(a.node.NodeName, a.getAwsTemplatePath(), a.awsS3Client, false); err != nil {
		return err
	}

	// destroy node [using terraform]
	if err := utils.DestroyNode(a.node.NodeName, values); err != nil {
		return err
	}

	return nil
}

func NewAwsProviderClient(node AWSNodeConfig, cpd common.CommonProviderData, apc AwsProviderConfig) (common.ProviderClient, error) {
	awsS3Client, err := awss3.NewAwsS3Client(apc.AccessKey, apc.AccessSecret, apc.AccountName)
	if err != nil {
		fmt.Println(utils.ColorText(err.Error(), 1))
		return nil, err
	}

	return AwsClient{
		node:        node,
		awsS3Client: awsS3Client,

		accessKey:    apc.AccessKey,
		accessSecret: apc.AccessSecret,
		accountName:  apc.AccountName,

		tfTemplates: cpd.TfTemplates,
		labels:      cpd.Labels,
		taints:      cpd.Taints,
	}, nil
}
