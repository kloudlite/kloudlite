package client

import (
	"errors"
	"fmt"
	"os"
	"path"

	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/adrg/xdg"
	"sigs.k8s.io/yaml"
)

type Env struct {
	Name          string `yaml:"name"`
	IsEnvironment bool   `yaml:"isEnvironment"`
}

type KLContext struct {
	AccountName  string          `yaml:"accountName"`
	DeviceName   string          `yaml:"deviceName"`
	Session      string          `yaml:"session"`
	KlFile       string          `yaml:"klFile"`
	DNS          []string        `yaml:"dns"`
	ClusterName  string          `yaml:"clusterName"`
	SelectedEnvs map[string]*Env `yaml:"selectedEnvs"`
}

func (f *KLContext) GetCookieString() string {
	return fmt.Sprintf("kloudlite-account=%s;kloudlite-cluster=%s;hotspot-session=%s", f.AccountName, f.ClusterName, f.Session)
}

func GetConfigFolder() (configFolder string, err error) {

	dirName := xdg.CacheHome

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

	filePath := path.Join(configPath, "config")

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

	cfile := path.Join(filePath, "config")

	err = os.WriteFile(cfile, file, 0644)
	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err = fn.ExecCmd(fmt.Sprintf("chown %s %s", usr, cfile),
			false); err != nil {
			return err
		}
	}

	if err != nil {
		fn.PrintError(err)
	}

	return err
}
