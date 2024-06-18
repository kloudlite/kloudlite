package client

import (
	"encoding/json"
	"fmt"
	"os"

	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type KLFileType struct {
	Version    string   `json:"version" yaml:"version"`
	DefaultEnv string   `json:"defaultEnv" yaml:"defaultEnv"`
	Packages   []string `json:"packages" yaml:"packages"`

	EnvVars EnvVars `json:"envVars" yaml:"envVars"`
	Mounts  Mounts  `json:"mounts" yaml:"mounts"`

	InitScripts []string `json:"initScripts" yaml:"initScripts"`
	AccountName string   `json:"accountName" yaml:"accountName"`
}

func (k *KLFileType) ToJson() ([]byte, error) {
	if k == nil {
		return nil, fmt.Errorf("kl file is nil")
	}

	return json.Marshal(*k)
}

func (k *KLFileType) ParseJson(b []byte) error {
	if k == nil {
		return fmt.Errorf("kl file is nil")
	}

	return json.Unmarshal(b, k)
}

const (
	defaultKLFile = "kl.yml"
)

func GetConfigPath() string {
	klfilepath := os.Getenv("KLCONFIG_PATH")
	if klfilepath != "" {
		return klfilepath
	}
	return defaultKLFile
}

func WriteKLFile(fileObj KLFileType) error {
	if err := confighandler.WriteConfig(GetConfigPath(), fileObj, 0644); err != nil {
		fn.PrintError(err)
		return err
	}

	return nil
}

func GetKlFile(filePath string) (*KLFileType, error) {
	if filePath == "" {
		s := GetConfigPath()
		filePath = s
	}

	klfile, err := confighandler.ReadConfig[KLFileType](filePath)
	if err != nil {
		return nil, err
	}

	return klfile, nil
}
