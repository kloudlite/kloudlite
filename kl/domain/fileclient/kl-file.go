package fileclient

import (
	"os"

	"github.com/kloudlite/kl/domain/envclient"
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
	TeamName string `json:"teamName" yaml:"teamName"`
}

const (
	defaultKLFile = "kl.yml"
)

func getConfigPath() string {
	klfilepath := os.Getenv("KLCONFIG_PATH")
	if klfilepath != "" {
		return klfilepath
	}

	if envclient.InsideBox() {
		s, err := envclient.GetWorkspacePath()
		if err != nil {
			return defaultKLFile
		}
		return s
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
	klfile, err := c.getKlFile(filePath)
	if err != nil {
		return nil, functions.NewE(err)
	}
	if klfile.TeamName == "" {
		return nil, functions.Error("kl file is not valid, teamName is not set properly. You can re-initialize kl file by running \"kl init\" command.")
	}
	if klfile.DefaultEnv == "" {
		return nil, functions.Error("kl file is not valid, defaultEnv is not set properly. You can re-intialize kl file by running \"kl init\" command.")
	}
	return klfile, nil
}

func (c *fclient) getKlFile(filePath string) (*KLFileType, error) {
	if filePath == "" {
		s := getConfigPath()
		filePath = s
	}

	klfile, err := confighandler.ReadConfig[KLFileType](filePath)
	if err != nil {
		return nil, functions.NewE(err, "failed to read klfile")
	}

	return klfile, nil
}
