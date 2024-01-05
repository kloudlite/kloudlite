package client

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"

	"sigs.k8s.io/yaml"
)

const (
	ConfigFileName string = "kl-session.yaml"
)

type Env struct {
	Name     string `json:"name"`
	TargetNs string `json:"targetNamespace"`
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
	homePath := ""

	// fetching homePath
	{
		if euid := os.Geteuid(); euid == 0 {
			username, ok := os.LookupEnv("SUDO_USER")
			if !ok {
				return "", errors.New("something went wrong")
			}

			oldPwd, err := os.Getwd()
			if err != nil {
				return "", err
			}

			sp := strings.Split(oldPwd, "/")

			for i := range sp {
				if sp[i] == username {
					homePath = path.Join("/", path.Join(sp[:i+1]...))
				}
			}
		} else {
			userHome, ok := os.LookupEnv("HOME")
			if !ok {
				return "", errors.New("something went wrong")
			}

			homePath = userHome
		}
	}

	configPath := path.Join(homePath, ".cache", ".kl")

	// ensuring the dir is present
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return "", err
	}

	// ensuring user permission on created dir
	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err = fn.ExecCmd(
			fmt.Sprintf("chown %s %s", usr, configPath), nil, false,
		); err != nil {
			return "", err
		}
	}

	return configPath, nil
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
