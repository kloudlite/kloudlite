package client

import (
	"errors"
	"fmt"
	"os"
	"path"

	fn "github.com/kloudlite/kl/pkg/functions"

	"sigs.k8s.io/yaml"
)

const (
	ConfigFileName string = "kl-session.yaml"
)

type Env struct {
	Name          string `json:"name"`
	IsEnvironment bool   `json:"isEnvironment"`
}

type KLContext struct {
	AccountName  string          `json:"accountName"`
	DeviceName   string          `json:"deviceName"`
	Session      string          `json:"session"`
	KlFile       string          `json:"klFile"`
	DNS          []string        `json:"dns"`
	ClusterName  string          `json:"clusterName"`
	SelectedEnvs map[string]*Env `json:"selectedEnvs"`
}

func (f *KLContext) GetCookieString() string {
	return fmt.Sprintf("kloudlite-account=%s;kloudlite-cluster=%s;hotspot-session=%s", f.AccountName, f.ClusterName, f.Session)
}

func GetConfigFolder() (configFolder string, err error) {

	var dirName string
	dirName, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		dirName, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}

	if dirName == "/root" {
		dirName, ok = os.LookupEnv("SUDO_USER")
		if !ok {
			return "", errors.New("something went wrong")
		}

		dirName = "/home/" + dirName
	}

	configFolder = fmt.Sprintf("%s/.kl", dirName)
	if _, err := os.Stat(configFolder); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(configFolder, os.ModePerm)
		if err != nil {
			fn.PrintError(err)
		}
	}
	return configFolder, nil
}

func GetContextFile() (*KLContext, error) {
	configPath, err := GetConfigFolder()
	if err != nil {
		return nil, err
	}

	filePath := path.Join(configPath, ConfigFileName)

	if _, er := os.Stat(filePath); errors.Is(er, os.ErrNotExist) {
		er := os.MkdirAll(path.Dir(filePath), os.ModePerm)
		if er != nil {
			return nil, er
		}

		ctx, _ := yaml.Marshal(KLContext{})

		e := os.WriteFile(filePath, ctx, os.ModePerm)
		if e != nil {
			return nil, e
		}
	}

	file, err := os.ReadFile(filePath)
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

func WriteContextFile(fileObj KLContext) error {
	filePath, err := GetConfigFolder()
	if err != nil {
		return err
	}

	file, err := yaml.Marshal(fileObj)
	if err != nil {
		fn.PrintError(err)
		return nil
	}

	cfile := path.Join(filePath, ConfigFileName)

	err = os.WriteFile(cfile, file, 0644)
	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err = fn.ExecCmd(
			fmt.Sprintf("chown %s %s", usr, cfile), nil, false,
		); err != nil {
			return err
		}
	}

	if err != nil {
		fn.PrintError(err)
	}

	return err
}
