package fileclient

import (
	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func getConfigPath() string {
	return "platform-config.yml"
}

type KlConfig struct {
	BaseDomain string `json:"baseDomain" yaml:"baseDomain"`
}

type ConfigFile struct {
	Version string `json:"version" yaml:"version"`

	KlConfig *KlConfig `json:"klConfig" yaml:"klConfig"`
	K3sArgs  []string  `json:"k3sConfig" yaml:"k3sArgs"`
}

func (f *fclient) WriteConfigFile(fileObj ConfigFile) error {
	if err := confighandler.WriteConfig(getConfigPath(), fileObj, 0644); err != nil {
		fn.PrintError(err)
		return fn.NewE(err)
	}

	return nil

}

func (f *fclient) GetConfigFile() (*ConfigFile, error) {
	confFile, err := f.getConfFile()
	if err != nil {
		return nil, functions.NewE(err)
	}
	return confFile, nil

}

func (f *fclient) getConfFile() (*ConfigFile, error) {
	s := getConfigPath()
	filePath := s

	confFile, err := confighandler.ReadConfig[ConfigFile](filePath)
	if err != nil {
		return nil, functions.NewE(err, "failed to read config file")
	}

	return confFile, nil
}
