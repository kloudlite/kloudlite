package server

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
	"kloudlite.io/cmd/internal/common"
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
	Type string `json:"type"`
	Name string `json:"Name"`
}

type MountType struct {
	MountBasePath string      `yaml:"mountBasePath" json:"mountBasePath"`
	Mounts        []FileEntry `json:"mounts"`
}

type KLFileType struct {
	Version   string    `json:"version"`
	Name      string    `json:"name"`
	Mres      []ResType `json:"mres"`
	Configs   []ResType `json:"configs"`
	Secrets   []ResType `json:"secrets"`
	Env       []EnvType `json:"env"`
	Ports     []string  `json:"ports"`
	FileMount MountType `yaml:"fileMount"`
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
		common.PrintError(err)
		return nil
	}

	err = ioutil.WriteFile(GetConfigPath(), file, 0644)
	if err != nil {
		common.PrintError(err)
	}

	return err
}

func GetKlFile(filePath *string) (*KLFileType, error) {
	if filePath == nil {
		s := GetConfigPath()
		filePath = &s
	}

	file, err := ioutil.ReadFile(*filePath)
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
