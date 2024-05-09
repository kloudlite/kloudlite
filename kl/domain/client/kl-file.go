package client

import (
	"os"

	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type CSType string

const (
	ConfigType CSType = "config"
	SecretType CSType = "secret"
)

type ResEnvType struct {
	Name   *string `json:"name,omitempty" yaml:"name,omitempty"`
	Key    string  `json:"key"`
	RefKey string  `json:"refKey"`
}

type EnvType struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

type ResType struct {
	Name string       `json:"name"`
	Env  []ResEnvType `json:"env"`
}

type FileEntry struct {
	Path string `json:"path"`
	Type CSType `json:"type"`
	Name string `json:"Name"`
	Key  string `json:"key"`
}

type MountType struct {
	MountBasePath string      `json:"mountBasePath"`
	Mounts        []FileEntry `json:"mounts"`
}

type KLFileType struct {
	Version    string    `json:"version"`
	DefaultEnv string    `json:"defaultEnv"`
	Packages   []string  `json:"packages"`
	Mres       []ResType `json:"mres"`
	Configs    []ResType `json:"configs"`
	Secrets    []ResType `json:"secrets"`
	Env        []EnvType `json:"env"`
	FileMount  MountType `json:"fileMount"`
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

	// file, err := yaml.Marshal(fileObj)
	// if err != nil {
	// 	fn.PrintError(err)
	// 	return nil
	// }
	//
	// writeContent := fmt.Sprint("# To generate this config file please visit ", constants.ServerURL, "\n\n", string(file))

	// err = os.WriteFile(GetConfigPath(), []byte(writeContent), 0644)
	// if err != nil {
	// 	fn.PrintError(err)
	// }

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
