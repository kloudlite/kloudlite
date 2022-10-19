package server

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/kloudlite/kl/lib/common"
	"gopkg.in/yaml.v2"
)

type KLContext struct {
	ProjectId string
	AccountId string
	DeviceId  string
	Session   string
	KlFile    string
	DNS       []string
}

func getConfigFolder() (configFolder string, err error) {
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
			common.PrintError(err)
		}
	}
	return configFolder, nil
}

func ActiveDns() ([]string, error) {

	file, err := GetContextFile()

	if err != nil {
		return nil, err
	}

	// if len(file.DNS) == 0 {
	// 	return nil,
	// 		errors.New("no active dns found")
	// }

	return file.DNS, nil
}

func SetActiveDns(dns []string) error {
	file, err := GetContextFile()
	if err != nil {
		return err
	}
	file.DNS = dns
	return WriteContextFile(*file)
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

	cfile := path.Join(filePath, "config")

	err = ioutil.WriteFile(cfile, file, 0644)
	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err = execCmd(fmt.Sprintf("chown %s %s", usr, cfile),
			false); err != nil {
			return err
		}
	}

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

func execCmd(cmdString string, verbose bool) error {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	if verbose {
		common.Log("[#] " + strings.Join(cmdArr, " "))
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	// s.Start()
	err = cmd.Run()
	// s.Stop()
	return err
}
