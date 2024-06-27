package client

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"

	"sigs.k8s.io/yaml"
)

const (
	SessionFileName   string = "kl-session.yaml"
	MainCtxFileName   string = "kl-main-contexts.yaml"
	ExtraDataFileName string = "kl-extra-data.yaml"
	DeviceFileName    string = "kl-device.yaml"
	CompleteFileName  string = "kl-completion"
)

type Env struct {
	Name        string `json:"name"`
	TargetNs    string `json:"targetNamespace"`
	SSHPort     int    `json:"sshPort"`
	ClusterName string `json:"clusterName"`
}

type Session struct {
	Session string `json:"session"`
}

type MainContext struct {
	AccountName string `json:"accountName"`
	ClusterName string `json:"clusterName"`
}

type DeviceContext struct {
	DeviceName string `json:"deviceName"`

	// PrivateKey []byte `json:"privateKey"`
	DeviceIp  net.IP   `json:"deviceIp"`
	DeviceDns []string `json:"deviceDns"`

	SearchDomain string `json:"searchDomain"`
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
	BaseUrl      string          `json:"baseUrl"`
	SelectedEnvs map[string]*Env `json:"selectedEnvs"`
	// DeviceDns    string          `json:"deviceDns"`
	// backupDns
	BackupDns    []string `json:"dns"`
	Loading      bool     `json:"loading"`
	VpnConnected bool     `json:"vpnConnected"`

	// TODO: don't have any idea about this field, needs to remove if not required
	ActiveCluster string `json:"activeCluster"`
	// SearchDomainAdded bool   `json:"searchDomainAdded"`
	// DnsAdded          bool            `json:"dnsAdded"`
	// DnsValues         []string        `json:"dnsValues"`
}

// func GetDeviceDns() (string, error) {
// 	extraData, err := GetExtraData()
// 	if err != nil {
// 		return "", functions.NewE(err)
// 	}

// 	return extraData.DeviceDns, nil
// }

// func SetDeviceDns(dns string) error {
// 	extraData, err := GetExtraData()
// 	if err != nil {
// 		return functions.NewE(err)
// 	}

// 	extraData.DeviceDns = dns

// 	file, err := yaml.Marshal(extraData)
// 	if err != nil {
// 		return functions.NewE(err)
// 	}

// 	return writeOnUserScope(ExtraDataFileName, file)
// }

func GetActiveCluster() (string, error) {
	extraData, err := GetExtraData()
	if err != nil {
		return "", functions.NewE(err)
	}

	return extraData.ActiveCluster, nil
}

func SetActiveCluster(devCluster string) error {
	extraData, err := GetExtraData()
	if err != nil {
		return functions.NewE(err)
	}

	extraData.ActiveCluster = devCluster

	file, err := yaml.Marshal(extraData)
	if err != nil {
		return functions.NewE(err)
	}

	return writeOnUserScope(ExtraDataFileName, file)
}

func GetDns() ([]string, error) {
	extraData, err := GetExtraData()
	if err != nil {
		return nil, functions.NewE(err)
	}

	return extraData.BackupDns, nil
}

func SetDns(dns []string) error {
	extraData, err := GetExtraData()
	if err != nil {
		return functions.NewE(err)
	}

	extraData.BackupDns = dns

	file, err := yaml.Marshal(extraData)
	if err != nil {
		return functions.NewE(err)
	}

	return writeOnUserScope(ExtraDataFileName, file)
}

func GetUserHomeDir() (string, error) {
	if runtime.GOOS == "windows" {
		return xdg.Home, nil
	}

	if euid := os.Geteuid(); euid == 0 {
		username, ok := os.LookupEnv("SUDO_USER")
		if !ok {
			return "", functions.Error("failed to get sudo user name")
		}

		oldPwd, err := os.Getwd()
		if err != nil {
			return "", functions.NewE(err)
		}

		sp := strings.Split(oldPwd, "/")

		for i := range sp {
			if sp[i] == username {
				return path.Join("/", path.Join(sp[:i+1]...)), nil
			}
		}

		return "", functions.Error("failed to get home path of sudo user")
	}

	userHome, ok := os.LookupEnv("HOME")
	if !ok {
		return "", functions.Error("failed to get home path of user")
	}

	return userHome, nil
}

func GetConfigFolder() (configFolder string, err error) {
	homePath, err := GetUserHomeDir()
	if err != nil {
		return "", functions.NewE(err)
	}

	configPath := path.Join(homePath, ".cache", ".kl")

	// ensuring the dir is present
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return "", functions.NewE(err)
	}

	// ensuring user permission on created dir
	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err = fn.ExecCmd(
			fmt.Sprintf("chown %s %s", usr, configPath), nil, false,
		); err != nil {
			return "", functions.NewE(err)
		}
	}

	return configPath, nil
}

func SetAccountToMainCtx(aName string) error {
	c, err := GetMainCtx()
	if err != nil {
		return functions.NewE(err)
	}

	c.AccountName = aName
	file, err := yaml.Marshal(c)
	if err != nil {
		return functions.NewE(err)
	}

	return writeOnUserScope(MainCtxFileName, file)
}

func SetClusterToMainCtx(cName string) error {
	c, err := GetMainCtx()
	if err != nil {
		return functions.NewE(err)
	}

	c.ClusterName = cName
	file, err := yaml.Marshal(c)
	if err != nil {
		return functions.NewE(err)
	}

	return writeOnUserScope(MainCtxFileName, file)
}

func GetMainCtx() (*MainContext, error) {
	file, err := ReadFile(MainCtxFileName)
	contexts := MainContext{}

	// need to check if file exists
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {

			b, err := yaml.Marshal(contexts)
			if err != nil {
				return nil, functions.NewE(err)
			}

			if err := writeOnUserScope(MainCtxFileName, b); err != nil {
				return nil, functions.NewE(err)
			}

		}
	}

	if err = yaml.Unmarshal(file, &contexts); err != nil {
		return nil, functions.NewE(err)
	}

	return &contexts, nil
}

func DeleteDeviceContext(dName string) error {
	if dName == "" {
		return fmt.Errorf("device Name is required")
	}

	c, err := GetDeviceContext()

	if err != nil {
		return functions.NewE(err)
	}

	if c.DeviceName != dName {
		return fmt.Errorf("device %s not found", dName)
	}

	c.DeviceName = ""

	b, err := yaml.Marshal(c)
	if err != nil {
		return functions.NewE(err)
	}

	return writeOnUserScope(DeviceFileName, b)
}

func WriteDeviceContext(dc *DeviceContext) error {
	file, err := yaml.Marshal(dc)

	if err != nil {
		return functions.NewE(err)
	}

	return writeOnUserScope(DeviceFileName, file)
}

func WriteCompletionContext() (io.Writer, error) {
	dir, err := GetConfigFolder()
	if err != nil {
		return nil, functions.NewE(err)
	}

	filePath := path.Join(dir, CompleteFileName)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, functions.NewE(err)
	}

	return file, nil
}

func GetCompletionContext() (string, error) {
	dir, err := GetConfigFolder()
	if err != nil {
		return "", functions.NewE(err)
	}

	filePath := path.Join(dir, CompleteFileName)
	return filePath, nil
}

func GetDeviceContext() (*DeviceContext, error) {
	file, err := ReadFile(DeviceFileName)
	contexts := DeviceContext{}

	// need to check if file exists
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {

			b, err := yaml.Marshal(contexts)
			if err != nil {
				return nil, functions.NewE(err)
			}

			if err := writeOnUserScope(DeviceFileName, b); err != nil {
				return nil, functions.NewE(err)
			}

		}
	}

	if err = yaml.Unmarshal(file, &contexts); err != nil {
		return nil, functions.NewE(err)
	}

	return &contexts, nil
}

func SaveBaseURL(url string) error {
	extraData, err := GetExtraData()
	if err != nil {
		return functions.NewE(err)
	}

	extraData.BaseUrl = url
	file, err := yaml.Marshal(extraData)
	if err != nil {
		return functions.NewE(err)
	}

	return writeOnUserScope(ExtraDataFileName, file)
}

func GetBaseURL() (string, error) {
	extraData, err := GetExtraData()
	if err != nil {
		return "", functions.NewE(err)
	}

	return extraData.BaseUrl, nil
}

func SaveExtraData(extraData *ExtraData) error {
	file, err := yaml.Marshal(extraData)
	if err != nil {
		return functions.NewE(err)
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
				return nil, functions.NewE(err)
			}

			if err := writeOnUserScope(ExtraDataFileName, b); err != nil {
				return nil, functions.NewE(err)
			}
		}

		return &extraData, nil
	}

	if err = yaml.Unmarshal(file, &extraData); err != nil {
		return nil, functions.NewE(err)
	}

	return &extraData, nil
}

func GetCookieString(options ...fn.Option) (string, error) {

	accName := fn.GetOption(options, "accountName")

	session, err := GetAuthSession()
	if err != nil {
		return "", functions.NewE(err)
	}

	if session == "" {
		return "", fmt.Errorf("no session found")
	}

	if accName != "" {
		return fmt.Sprintf("kloudlite-account=%s;hotspot-session=%s", accName, session), nil
	}

	c, err := GetMainCtx()
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
				return "", functions.NewE(err)
			}

			if err := writeOnUserScope(SessionFileName, b); err != nil {
				return "", functions.NewE(err)
			}
		}
	}

	if err = yaml.Unmarshal(file, &session); err != nil {
		return "", functions.NewE(err)
	}

	return session.Session, nil
}

func SaveAuthSession(session string) error {
	file, err := yaml.Marshal(Session{Session: session})
	if err != nil {
		return functions.NewE(err)
	}

	return writeOnUserScope(SessionFileName, file)
}

func writeOnUserScope(name string, data []byte) error {
	dir, err := GetConfigFolder()
	if err != nil {
		return functions.NewE(err)
	}

	if _, er := os.Stat(dir); errors.Is(er, os.ErrNotExist) {
		er := os.MkdirAll(dir, os.ModePerm)
		if er != nil {
			return er
		}
	}

	filePath := path.Join(dir, name)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return functions.NewE(err)
	}

	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err := fn.ExecCmd(
			fmt.Sprintf("chown %s %s", usr, filePath), nil, false,
		); err != nil {
			return functions.NewE(err)
		}
	}

	return nil
}

func ReadFile(name string) ([]byte, error) {
	dir, err := GetConfigFolder()
	if err != nil {
		return nil, functions.NewE(err)
	}

	filePath := path.Join(dir, name)

	if _, er := os.Stat(filePath); errors.Is(er, os.ErrNotExist) {
		return nil, fmt.Errorf("file not found")
	}

	file, err := os.ReadFile(filePath)

	if err != nil {
		return nil, functions.NewE(err)
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
		return functions.NewE(err)
	}

	extraData.Loading = loading

	return SaveExtraData(extraData)
}
