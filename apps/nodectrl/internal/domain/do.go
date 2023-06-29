package domain

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
	infraclient "kloudlite.io/pkg/infraClient"
)

type doConfig struct {
	Version  string `yaml:"version"`
	Action   string `yaml:"action"`
	Provider string `yaml:"provider"`
	Spec     struct {
		Provider struct {
			ApiToken  string `yaml:"apiToken"`
			AccountId string `yaml:"accountId"`
		} `yaml:"provider"`
		Node struct {
			Region  string `yaml:"region"`
			Size    string `yaml:"size"`
			NodeId  string `yaml:"nodeId"`
			ImageId string `yaml:"imageId"`
		} `yaml:"node"`
	} `yaml:"spec"`
}

func (d *domainI) doWithDO() error {

	out, err := base64.StdEncoding.DecodeString(d.env.Config)
	if err != nil {
		fmt.Println("here")
		return err
	}

	var doConf doConfig
	e := yaml.Unmarshal(out, &doConf)
	if e != nil {
		return e
	}
	klConf, err := d.getKlConf()
	if err != nil {
		fmt.Println("here")
		return err
	}

	labels := map[string]string{}
	if e := json.Unmarshal([]byte(d.env.Labels), &labels); e != nil {
		fmt.Println(e)
	}

	taints := []string{}
	if e := json.Unmarshal([]byte(d.env.Taints), &taints); e != nil {
		fmt.Println(e)
	}

	doProvider := infraclient.NewDOProvider(infraclient.DoProvider{
		ApiToken:  doConf.Spec.Provider.ApiToken,
		AccountId: doConf.Spec.Provider.AccountId,
	}, infraclient.DoProviderEnv{
		StorePath:   klConf.Values.StorePath,
		TfTemplates: klConf.Values.TfTemplates,
		Secrets:     klConf.Values.Secrets,
		Labels:      labels,
		Taints:      taints,
		SSHPath:     klConf.Values.SSHPath,
	})

	doNode := infraclient.DoNode{
		Region:  doConf.Spec.Node.Region,
		Size:    doConf.Spec.Node.Size,
		NodeId:  doConf.Spec.Node.NodeId,
		ImageId: doConf.Spec.Node.ImageId,
	}

	// return nil

	switch doConf.Action {
	case "create":
		if err = doProvider.NewNode(doNode); err != nil {
			return err
		}

		if err = doProvider.AttachNode(doNode); err != nil {
			return err
		}

	case "delete":

		if err = doProvider.UnattachNode(doNode); err != nil {
			return err
		}

		if err = doProvider.DeleteNode(doNode); err != nil {
			return err
		}

	default:
		return errors.New("wrong action")
	}

	return nil
}
