package server

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
	"kloudlite.io/cmd/internal/common"
)

type KLContext struct {
	ProjectId string
	AccountId string
	Session   string
	KlFile    string
}

func WriteContextFile(fileObj KLContext) error {
	filePath, err := getConfigFolder()
	if err != nil {
		return err
	}

	file, err := yaml.Marshal(fileObj)
	if err != nil {
		common.PrintError(err)
		return nil
	}

	err = ioutil.WriteFile(path.Join(filePath, "config"), file, 0644)
	if err != nil {
		common.PrintError(err)
	}

	return err
}

func GetContextFile() (*KLContext, error) {
	configPath, err := getConfigFolder()
	if err != nil {
		return nil, err
	}

	filePath := path.Join(configPath, "config")

	if _, er := os.Stat(filePath); errors.Is(er, os.ErrNotExist) {
		er := os.MkdirAll(path.Dir(filePath), os.ModePerm)
		if er != nil {
			return nil, er
		}

		ctx, _ := yaml.Marshal(KLContext{})

		e := ioutil.WriteFile(filePath, ctx, os.ModePerm)
		if e != nil {
			return nil, e
		}
	}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	klfile := KLContext{}

	err = yaml.Unmarshal(file, &klfile)
	if err != nil {
		return nil, err
	}

	return &klfile, nil
}
