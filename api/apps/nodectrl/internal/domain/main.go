package domain

import (
	"encoding/base64"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
	"kloudlite.io/pkg/config"

	"go.uber.org/fx"
)

type domainI struct {
	env *Env
}

type KLConf struct {
	Version string `yaml:"version"`
	Values  struct {
		StorePath   string `yaml:"storePath"`
		TfTemplates string `yaml:"tfTemplatesPath"`
		Secrets     string `yaml:"secrets"`
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

// startJob implements Domain
func (d *domainI) StartJob() error {

	switch d.env.Provider {
	case "do":
		if err := d.doWithDO(); err != nil {
			return err
		}

	case "aws":
		if err := d.doWithAWS(); err != nil {
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
	Taints   string `env:"TAINTS" required:"true"`
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
