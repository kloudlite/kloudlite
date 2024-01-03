package client

import (
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"

	"sigs.k8s.io/yaml"
)

type ResEnvType struct {
	Name   *string `yaml:"name,omitempty" yaml:"name,omitempty"`
	Key    string  `yaml:"key"`
	RefKey string  `yaml:"refKey"`
}

type EnvType struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type ResType struct {
	Name string       `yaml:"name"`
	Env  []ResEnvType `yaml:"env"`
}

type FileEntry struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"`
	Name string `yaml:"Name"`
}

type MountType struct {
	MountBasePath string      `yaml:"mountBasePath" yaml:"mountBasePath"`
	Mounts        []FileEntry `yaml:"mounts"`
}

type KLFileType struct {
	Version   string    `yaml:"version"`
	Project   string    `yaml:"project"`
	Mres      []ResType `yaml:"mres"`
	Configs   []ResType `yaml:"configs"`
	Secrets   []ResType `yaml:"secrets"`
	Env       []EnvType `yaml:"env"`
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
		fn.PrintError(err)
		return nil
	}

	err = os.WriteFile(GetConfigPath(), file, 0644)
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
