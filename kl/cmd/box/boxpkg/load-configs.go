package boxpkg

import (
	mclient "github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
)

type EnvironmentVariable struct {
	Key   string `yaml:"key" json:"key"`
	Value string `yaml:"value" json:"value"`
}

type KLConfigType struct {
	Packages []string              `yaml:"packages" json:"packages"`
	EnvVars  []EnvironmentVariable `yaml:"envVars" json:"envVars"`
	Mounts   map[string]string     `yaml:"mounts" json:"mounts"`
}

func (*client) loadConfig(mm server.MountMap, envs map[string]string) (*KLConfigType, error) {
	kf, err := mclient.GetKlFile("")
	if err != nil {
		return nil, err
	}

	// read kl.yml into struct
	klConfig := &KLConfigType{
		Packages: kf.Packages,
	}

	kt, err := mclient.GetKlFile("")
	if err != nil {
		return nil, err
	}

	fm := map[string]string{}

	for _, fe := range kt.FileMount.Mounts {
		pth := fe.Path
		if pth == "" {
			pth = fe.Key
		}

		fm[pth] = mm[pth]
	}

	// return fm, nil

	var ev []EnvironmentVariable
	for k, v := range envs {
		ev = append(ev, EnvironmentVariable{k, v})
	}

	klConfig.EnvVars = ev
	if klConfig.EnvVars == nil {
		klConfig.EnvVars = []EnvironmentVariable{}
	}
	klConfig.Mounts = fm

	return klConfig, nil
}
