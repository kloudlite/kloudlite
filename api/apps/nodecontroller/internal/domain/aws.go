package domain

import (
	"encoding/base64"
	"errors"

	"gopkg.in/yaml.v3"
	infraclient "kloudlite.io/pkg/infraClient"
)

type awsConfig struct {
	Version  string `yaml:"version"`
	Action   string `yaml:"action"`
	Provider string `yaml:"provider"`
	Spec     struct {
		Provider struct {
			AccessKey    string `yaml:"accessKey"`
			AccessSecret string `yaml:"accessSecret"`
			AccountId    string `yaml:"accountId"`
		} `yaml:"provider"`
		Node struct {
			Region       string `yaml:"region"`
			InstanceType string `yaml:"instanceType"`
			NodeId       string `yaml:"nodeId"`
			VPC          string `yaml:"vpc"`
		} `yaml:"node"`
	} `yaml:"spec"`
}

func (d *domainI) doWithAWS() error {

	out, err := base64.StdEncoding.DecodeString(d.env.Config)
	if err != nil {
		return err
	}
	var awsConf awsConfig
	e := yaml.Unmarshal(out, &awsConf)
	if e != nil {
		return e
	}
	klConf, err := d.getKlConf()
	if err != nil {
		return err
	}

	awsProvider := infraclient.NewAWSProvider(infraclient.AWSProvider{
		AccessKey:    awsConf.Spec.Provider.AccessKey,
		AccessSecret: awsConf.Spec.Provider.AccessSecret,
		AccountId:    awsConf.Spec.Provider.AccountId,
	}, infraclient.AWSProviderEnv{
		StorePath:   klConf.Values.StorePath,
		TfTemplates: klConf.Values.TfTemplates,
		SSHPath:     klConf.Values.SSHPath,
	})

	awsNode := infraclient.AWSNode{
		NodeId:       awsConf.Spec.Node.NodeId,
		Region:       awsConf.Spec.Node.Region,
		InstanceType: awsConf.Spec.Node.InstanceType,
		VPC:          awsConf.Spec.Node.VPC,
	}

	// return nil

	switch awsConf.Action {
	case "create":
		err = awsProvider.NewNode(awsNode)
		if err != nil {
			return err
		}

	case "delete":
		err = awsProvider.DeleteNode(awsNode)

		if err != nil {
			return err
		}

	default:
		return errors.New("wrong action")
	}

	return nil
}
