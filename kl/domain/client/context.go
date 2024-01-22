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
	ConfigFileName          string = "kl-session.yaml"
	SessionFileName         string = "kl-session.yaml"
	AccountContextsFileName string = "kl-account-contexts.yaml"
	ExtraDataFileName       string = "kl-extra-data.yaml"
	InfraContextsFileName   string = "kl-infra-contexts.yaml"
	DeviceFileName          string = "kl-device.yaml"
)

type Env struct {
	Name     string `json:"name"`
	TargetNs string `json:"targetNamespace"`
}

type Session struct {
	Session string `json:"session"`
}

type AccountContext struct {
	AccountName string `json:"accountName"`
}

type DeviceContext struct {
	DeviceName string `json:"deviceName"`
}

type InfraContext struct {
	Name        string `json:"name"`
	AccountName string `json:"accountName"`
	ClusterName string `json:"ClusterName"`
	DeviceName  string `json:"deviceName"`
}

type InfraContexts struct {
	InfraContexts map[string]*InfraContext `json:"infraContexts"`
	ActiveContext string                   `json:"activeContext"`
}

type ExtraData struct {
	SelectedEnvs map[string]*Env `json:"selectedEnvs"`
	DNS          []string        `json:"dns"`
	Loading      bool            `json:"loading"`
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

func GetActiveInfraContext() (*InfraContext, error) {
	c, err := GetInfraContexts()
	if err != nil {
		return nil, err
	}

	if c.ActiveContext == "" {
		return &InfraContext{}, nil
	}

	if c.InfraContexts == nil {
		c.InfraContexts = map[string]*InfraContext{}
	}

	ctx, ok := c.InfraContexts[c.ActiveContext]
	if !ok {
		return &InfraContext{}, nil
	}

	return ctx, nil
}

func SetActiveInfraContext(name string) error {
	file, err := GetInfraContexts()

	if err != nil {
		return err
	}

	file.ActiveContext = name

	b, err := yaml.Marshal(file)

	if err != nil {
		return err
	}

	return writeOnUserScope(InfraContextsFileName, b)
}

func DeleteInfraContext(name string) error {
	if name == "" {
		return fmt.Errorf("context name is required")
	}

	c, err := GetInfraContexts()

	if err != nil {
		return err
	}

	if _, ok := c.InfraContexts[name]; !ok {
		return fmt.Errorf("context %s not found", name)
	}

	delete(c.InfraContexts, name)

	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return writeOnUserScope(InfraContextsFileName, b)
}

func WriteInfraContextFile(fileObj InfraContext) error {
	c, err := GetInfraContexts()
	if err != nil {
		return err
	}
	if c.InfraContexts == nil {
		c.InfraContexts = map[string]*InfraContext{}
	}

	c.InfraContexts[fileObj.Name] = &fileObj

	file, err := yaml.Marshal(c)

	if err != nil {
		return err
	}

	return writeOnUserScope(InfraContextsFileName, file)
}

func GetInfraContexts() (*InfraContexts, error) {
	file, err := ReadFile(InfraContextsFileName)
	infraContexts := InfraContexts{}

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {

			b, err := yaml.Marshal(infraContexts)
			if err != nil {
				return nil, err
			}

			if err := writeOnUserScope(InfraContextsFileName, b); err != nil {
				return nil, err
			}

		}
	}

	if err = yaml.Unmarshal(file, &infraContexts); err != nil {
		return nil, err
	}

	return &infraContexts, nil
}

func GetInfraCookieString() (string, error) {
	session, err := GetAuthSession()
	if err != nil {
		return "", err
	}

	if session == "" {
		return "", fmt.Errorf("no session found")
	}

	c, err := GetInfraContexts()
	if err != nil {
		return fmt.Sprintf("hotspot-session=%s", session), nil
	}

	if c.ActiveContext == "" {
		return fmt.Sprintf("hotspot-session=%s", session), nil
	}

	ctx, ok := c.InfraContexts[c.ActiveContext]
	if !ok {
		return fmt.Sprintf("hotspot-session=%s", session), nil
	}

	return fmt.Sprintf("kloudlite-account=%s;hotspot-session=%s", ctx.AccountName, session), nil
}

func DeleteAccountContext(aName string) error {
	if aName == "" {
		return fmt.Errorf("Account Name is required")
	}

	c, err := GetAccountContext()

	if err != nil {
		return err
	}

	if c.AccountName != aName {
		return fmt.Errorf("Account %s not found", aName)
	}

	c.AccountName = ""

	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return writeOnUserScope(AccountContextsFileName, b)
}

func WriteAccountContext(aName string) error {
	c, err := GetAccountContext()
	if err != nil {
		return err
	}

	c.AccountName = aName

	file, err := yaml.Marshal(c)

	if err != nil {
		return err
	}

	return writeOnUserScope(AccountContextsFileName, file)
}

func GetAccountContext() (*AccountContext, error) {
	file, err := ReadFile(AccountContextsFileName)
	contexts := AccountContext{}

	// need to check if file exists
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {

			b, err := yaml.Marshal(contexts)
			if err != nil {
				return nil, err
			}

			if err := writeOnUserScope(AccountContextsFileName, b); err != nil {
				return nil, err
			}

		}
	}

	if err = yaml.Unmarshal(file, &contexts); err != nil {
		return nil, err
	}

	return &contexts, nil
}

func DeleteDeviceContext(dName string) error {
	if dName == "" {
		return fmt.Errorf("Device Name is required")
	}

	c, err := GetDeviceContext()

	if err != nil {
		return err
	}

	if c.DeviceName != dName {
		return fmt.Errorf("Device %s not found", dName)
	}

	c.DeviceName = ""

	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return writeOnUserScope(DeviceFileName, b)
}

func WriteDeviceContext(dName string) error {
	c, err := GetDeviceContext()
	if err != nil {
		return err
	}

	c.DeviceName = dName

	file, err := yaml.Marshal(c)

	if err != nil {
		return err
	}

	return writeOnUserScope(DeviceFileName, file)
}

func GetDeviceContext() (*DeviceContext, error) {
	file, err := ReadFile(DeviceFileName)
	contexts := DeviceContext{}

	// need to check if file exists
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {

			b, err := yaml.Marshal(contexts)
			if err != nil {
				return nil, err
			}

			if err := writeOnUserScope(DeviceFileName, b); err != nil {
				return nil, err
			}

		}
	}

	if err = yaml.Unmarshal(file, &contexts); err != nil {
		return nil, err
	}

	return &contexts, nil
}

func SaveExtraData(extraData *ExtraData) error {
	file, err := yaml.Marshal(extraData)
	if err != nil {
		return err
	}

	return writeOnUserScope(ExtraDataFileName, file)
}

func GetExtraData() (*ExtraData, error) {
	file, err := ReadFile(ExtraDataFileName)
	extraData := ExtraData{}
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			b, err := yaml.Marshal(extraData)

			if err != nil {
				return nil, err
			}

			if err := writeOnUserScope(ExtraDataFileName, b); err != nil {
				return nil, err
			}
		}

		return &extraData, nil
	}

	if err = yaml.Unmarshal(file, &extraData); err != nil {
		return nil, err
	}

	return &extraData, nil
}

func GetCookieString() (string, error) {
	session, err := GetAuthSession()
	if err != nil {
		return "", err
	}

	if session == "" {
		return "", fmt.Errorf("no session found")
	}

	c, err := GetAccountContext()
	if err != nil {
		return fmt.Sprintf("hotspot-session=%s", session), nil
	}

	if c.AccountName == "" {
		return fmt.Sprintf("hotspot-session=%s", session), nil
	}

	return fmt.Sprintf("kloudlite-account=%s;hotspot-session=%s", c.AccountName, session), nil
}

func GetAuthSession() (string, error) {
	file, err := ReadFile(SessionFileName)

	session := Session{}

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			b, err := yaml.Marshal(session)
			if err != nil {
				return "", err
			}

			if err := writeOnUserScope(SessionFileName, b); err != nil {
				return "", err
			}
		}
	}

	if err = yaml.Unmarshal(file, &session); err != nil {
		return "", err
	}

	return session.Session, nil
}

func SaveAuthSession(session string) error {
	file, err := yaml.Marshal(Session{Session: session})
	if err != nil {
		return err
	}

	return writeOnUserScope(SessionFileName, file)
}

func writeOnUserScope(name string, data []byte) error {
	dir, err := GetConfigFolder()
	if err != nil {
		return err
	}

	if _, er := os.Stat(dir); errors.Is(er, os.ErrNotExist) {
		er := os.MkdirAll(dir, os.ModePerm)
		if er != nil {
			return er
		}
	}

	filePath := path.Join(dir, name)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err := fn.ExecCmd(
			fmt.Sprintf("chown %s %s", usr, filePath), nil, false,
		); err != nil {
			return err
		}
	}

	return nil
}

func ReadFile(name string) ([]byte, error) {
	dir, err := GetConfigFolder()
	if err != nil {
		return nil, err
	}

	filePath := path.Join(dir, name)

	if _, er := os.Stat(filePath); errors.Is(er, os.ErrNotExist) {
		return nil, fmt.Errorf("file not found")
	}

	file, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func IsLoading() (bool, error) {
	extraData, err := GetExtraData()
	if err != nil {
		return false, err
	}

	return extraData.Loading, nil
}

func SetLoading(loading bool) error {
	extraData, err := GetExtraData()
	if err != nil {
		return err
	}

	extraData.Loading = loading

	return SaveExtraData(extraData)
}
