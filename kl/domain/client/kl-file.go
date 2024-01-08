package client

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/constants"
	fn "github.com/kloudlite/kl/pkg/functions"

	"sigs.k8s.io/yaml"
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
	Key   string `json:"key"`
	Value string `json:"value"`
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
	Version   string    `json:"version"`
	Project   string    `json:"project"`
	Mres      []ResType `json:"mres"`
	Configs   []ResType `json:"configs"`
	Secrets   []ResType `json:"secrets"`
	Env       []EnvType `json:"env"`
	FileMount MountType `json:"fileMount"`
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
	file, err := yaml.Marshal(fileObj)
	if err != nil {
		fn.PrintError(err)
		return nil
	}

	writeContent := fmt.Sprint("# To generate this config file please visit ", constants.ServerURL, "\n\n", string(file))

	err = os.WriteFile(GetConfigPath(), []byte(writeContent), 0644)
	if err != nil {
		fn.PrintError(err)
	}

	return err
}

func GetKlFile(filePath *string) (*KLFileType, error) {
	if filePath == nil {
		s := GetConfigPath()
		filePath = &s
	}

	file, err := os.ReadFile(*filePath)
	if err != nil {
		return nil, err
	}

	klfile := KLFileType{}

	err = yaml.Unmarshal(file, &klfile)
	if err != nil {
		return nil, err
	}

	return &klfile, nil
}
