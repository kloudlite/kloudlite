package devboxfile

import (
	"encoding/json"
	"fmt"
)

type KlConfig struct {
	Mounts map[string]string `yaml:"mounts" json:"mounts"`
	Dns    string            `yaml:"dns" json:"dns"`
	// InitScripts []string          `yaml:"initScripts" json:"initScripts"`
}

type DevboxConfig struct {
	Packages      []string          `yaml:"packages" json:"packages"`
	PackageHashes map[string]string `yaml:"packageHashes" json:"packageHashes"`
	Env           map[string]string `yaml:"env" json:"env"`
	KlConfig      KlConfig          `yaml:"kloudliteConfig" json:"kloudliteConfig"`
}

func (k *DevboxConfig) ToJson() ([]byte, error) {

	if k == nil {
		return nil, fmt.Errorf("kl file is nil")
	}

	return json.Marshal(*k)
}

func (k *DevboxConfig) ParseJson(b []byte) error {
	if k == nil {
		return fmt.Errorf("kl file is nil")
	}

	return json.Unmarshal(b, k)
}
