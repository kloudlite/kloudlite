package server

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
	"kloudlite.io/cmd/internal/common"
)

type ResEnvType struct {
	Key    string
	RefKey string
}

type EnvType struct {
	Key   string
	Value string
}

type ResType struct {
	Id   string
	Name string
	Env  []ResEnvType
}

type FileEntry struct {
	Path string
	Type string
	Ref  string
	Name string
}

type MountType struct {
	MountBasePath string `yaml:"mountBasePath"`
	Mounts        []FileEntry
}

type KLFileType struct {
	Version   string
	Name      string
	Mres      []ResType
	Configs   []ResType
	Secrets   []ResType
	Env       []EnvType
	Ports     []string
	FileMount MountType `yaml:"fileMount"`
}

func WriteKLFile(fileObj KLFileType) error {
	file, err := yaml.Marshal(fileObj)
	if err != nil {
		common.PrintError(err)
		return nil
	}

	err = ioutil.WriteFile(".kl.yml", file, 0644)
	if err != nil {
		common.PrintError(err)
	}

	return err
}

func GetKlFile(filePath *string) (*KLFileType, error) {
	if filePath == nil {
		s := ".kl.yml"
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
