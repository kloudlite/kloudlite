package domain

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
	"kloudlite.io/pkg/config"
	infraclient "kloudlite.io/pkg/infraClient"

	"go.uber.org/fx"
)

type domainI struct {
	env *Env
}

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

type KLConf struct {
	Version string `yaml:"version"`
	Values  struct {
		ServerUrl   string `yaml:"serverUrl"`
		SshKeyPath  string `yaml:"sshKeyPath"`
		StorePath   string `yaml:"storePath"`
		TfTemplates string `yaml:"tfTemplatesPath"`
		JoinToken   string `yaml:"joinToken"`
	} `yaml:"spec"`
}

func (d *domainI) getKlConf() (*KLConf, error) {
	out, err := base64.StdEncoding.DecodeString(d.env.KLConfig)
	if err != nil {
		fmt.Println("here")
		return nil, err
	}

	var klConf KLConf
	e := yaml.Unmarshal(out, &klConf)
	if e != nil {
		return nil, e
	}

	return &klConf, nil
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

	doProvider := infraclient.NewDOProvider(infraclient.DoProvider{
		ApiToken:  doConf.Spec.Provider.ApiToken,
		AccountId: doConf.Spec.Provider.AccountId,
	}, infraclient.ProviderEnv{
		ServerUrl:   klConf.Values.ServerUrl,
		SshKeyPath:  klConf.Values.SshKeyPath,
		StorePath:   klConf.Values.StorePath,
		TfTemplates: klConf.Values.TfTemplates,
		JoinToken:   klConf.Values.JoinToken,
		Labels:      labels,
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
		err = doProvider.NewNode(doNode)
		if err != nil {
			return err
		}
		err = doProvider.AttachNode(doNode)
		if err != nil {
			return err
		}

	case "delete":
		err = doProvider.DeleteNode(doNode)

		if err != nil {
			return err
		}

	default:
		return errors.New("wrong action")
	}

	return nil
}

// startJob implements Domain
func (d *domainI) StartJob() error {

	switch d.env.Provider {
	case "do":
		err := d.doWithDO()
		if err != nil {
			return err
		}

	default:
		return errors.New("this type of provider not suported")
	}
	return nil
}

func fxDomain(env *Env) Domain {
	return &domainI{
		env: env,
	}
}

type Env struct {
	Config   string `env:"NODE_CONFIG" required:"true"`
	Provider string `env:"PROVIDER" required:"true"`
	KLConfig string `env:"KL_CONFIG" required:"true"`
	Labels   string `env:"LABELS" required:"true"`
}

var Module = fx.Module(
	"domain",
	config.EnvFx[Env](),
	fx.Provide(fxDomain),
)

/*
main
	-> framework ()
	-> app ()
	-> domain (main logic)
*/
