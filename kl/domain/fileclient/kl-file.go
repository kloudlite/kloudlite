package fileclient

import (
	"os"

	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type KLFileType struct {
	Version    string   `json:"version" yaml:"version"`
	DefaultEnv string   `json:"defaultEnv" yaml:"defaultEnv"`
	Packages   []string `json:"packages" yaml:"packages"`

	EnvVars EnvVars `json:"envVars" yaml:"envVars"`
	Mounts  Mounts  `json:"mounts" yaml:"mounts"`
	Ports   []int   `json:"ports" yaml:"ports"`

	// InitScripts []string `json:"initScripts" yaml:"initScripts"`
	AccountName string `json:"accountName" yaml:"accountName"`
}

const (
	defaultKLFile = "kl.yml"
)

func getConfigPath() string {
	klfilepath := os.Getenv("KLCONFIG_PATH")
	if klfilepath != "" {
		return klfilepath
	}
	return defaultKLFile
}

func (c *fclient) WriteKLFile(fileObj KLFileType) error {
	if err := confighandler.WriteConfig(getConfigPath(), fileObj, 0644); err != nil {
		fn.PrintError(err)
		return functions.NewE(err)
	}

	return nil
}

func (c *fclient) GetKlFile(filePath string) (*KLFileType, error) {
	return c.getKlFile(filePath)
}

func (c *fclient) getKlFile(filePath string) (*KLFileType, error) {
	if filePath == "" {
		s := getConfigPath()
		filePath = s
	}

	klfile, err := confighandler.ReadConfig[KLFileType](filePath)
	if err != nil {
		return nil, functions.NewE(err)
	}

	return klfile, nil
}
