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
	ConfigFileName    string = "kl-session.yaml"
	SessionFileName   string = "kl-session.yaml"
	ContextsFileName  string = "kl-contexts.yaml"
	ExtraDataFileName string = "kl-extra-data.yaml"
)

type Env struct {
	Name     string `json:"name"`
	TargetNs string `json:"targetNamespace"`
}

type Session struct {
	Session string `json:"session"`
}

type Context struct {
	Name        string `json:"name"`
	AccountName string `json:"accountName"`
	DeviceName  string `json:"deviceName"`
	ClusterName string `json:"clusterName"`
}

type Contexts struct {
	Contexts      map[string]*Context `json:"contexts"`
	ActiveContext string              `json:"activeContext"`
}

type ExtraData struct {
	SelectedEnvs map[string]*Env `json:"selectedEnvs"`
	DNS          []string        `json:"dns"`
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

func GetContextFile() (*Context, error) {
	c, err := GetContexts()
	if err != nil {
		return nil, err
	}

	if c.ActiveContext == "" {
		return &Context{}, nil
	}

	if c.Contexts == nil {
		c.Contexts = map[string]*Context{}
	}

	ctx, ok := c.Contexts[c.ActiveContext]
	if !ok {
		return &Context{}, nil
	}

	return ctx, nil
}

func SetActiveContext(name string) error {
	file, err := GetContexts()

	if err != nil {
		return err
	}

	file.ActiveContext = name

	b, err := yaml.Marshal(file)

	if err != nil {
		return err
	}

	return writeOnUserScope(ContextsFileName, b)
}

func DeleteContext(name string) error {
	if name == "" {
		return fmt.Errorf("context name is required")
	}

	c, err := GetContexts()

	if err != nil {
		return err
	}

	if _, ok := c.Contexts[name]; !ok {
		return fmt.Errorf("context %s not found", name)
	}

	delete(c.Contexts, name)

	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return writeOnUserScope(ContextsFileName, b)
}

func WriteContextFile(fileObj Context) error {
	c, err := GetContexts()
	if err != nil {
		return err
	}
	if c.Contexts == nil {
		c.Contexts = map[string]*Context{}
	}

	c.Contexts[fileObj.Name] = &fileObj

	file, err := yaml.Marshal(c)

	if err != nil {
		return err
	}

	return writeOnUserScope(ContextsFileName, file)
}

func GetContexts() (*Contexts, error) {
	file, err := ReadFile(ContextsFileName)
	contexts := Contexts{}

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {

			b, err := yaml.Marshal(contexts)
			if err != nil {
				return nil, err
			}

			if err := writeOnUserScope(ContextsFileName, b); err != nil {
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

	c, err := GetContexts()
	if err != nil {
		return fmt.Sprintf("hotspot-session=%s", session), nil
	}

	if c.ActiveContext == "" {
		return fmt.Sprintf("hotspot-session=%s", session), nil
	}

	ctx, ok := c.Contexts[c.ActiveContext]
	if !ok {
		return fmt.Sprintf("hotspot-session=%s", session), nil
	}

	return fmt.Sprintf("kloudlite-account=%s;kloudlite-cluster=%s;hotspot-session=%s", ctx.AccountName, ctx.ClusterName, session), nil
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
