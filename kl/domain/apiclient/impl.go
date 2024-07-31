package apiclient

import (
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type apiClient struct {
	fc fileclient.FileClient
}

type ApiClient interface {
	ListAccounts() ([]Account, error)

	ListApps(accountName string, envName string) ([]App, error)
	InterceptApp(app *App, status bool, ports []AppPort, envName string, options ...fn.Option) (err error)

	CreateRemoteLogin() (loginId string, err error)
	GetCurrentUser() (*User, error)
	Login(loginId string) error

	ListConfigs(accountName string, envName string) ([]Config, error)
	GetConfig(accountName string, envName string, configName string) (*Config, error)

	GetVPNDevice(accountName string, devName string) (*Device, error)
	CheckDeviceStatus() bool
	GetAccVPNConfig(account string) (*fileclient.AccountVpnConfig, error)

	ListEnvs(accountName string) ([]Env, error)
	GetEnvironment(accountName, envName string) (*Env, error)
	EnsureEnv() (*fileclient.Env, error)
	GetLoadMaps() (map[string]string, MountMap, error)

	ListMreses(accountName string, envName string) ([]Mres, error)
	ListMresKeys(accountName, envName, importedManagedResource string) ([]string, error)
	GetMresConfigValues(accountName string) (map[string]string, error)

	ListSecrets(accountName string, envName string) ([]Secret, error)
	GetSecret(accountName string, secretName string) (*Secret, error)

	RemoveAllIntercepts(options ...fn.Option) error
}

func New() (ApiClient, error) {
	fc, err := fileclient.New()
	if err != nil {
		return nil, fn.NewE(err)
	}
	return &apiClient{
		fc: fc,
	}, nil
}
