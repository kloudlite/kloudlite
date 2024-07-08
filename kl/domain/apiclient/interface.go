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

	ListApps(options ...fn.Option) ([]App, error)
	InterceptApp(app *App, status bool, ports []AppPort, options ...fn.Option) error

	CreateRemoteLogin() (loginId string, err error)
	GetCurrentUser() (*User, error)
	Login(loginId string) error

	ListConfigs(accountName string, envName string) ([]Config, error)
	GetConfig(options ...fn.Option) (*Config, error)

	GetVPNDevice(devName string, options ...fn.Option) (*Device, error)
	CheckDeviceStatus() bool
	GetAccVPNConfig(account string) (*fileclient.AccountVpnConfig, error)

	ListEnvs(options ...fn.Option) ([]Env, error)
	GetEnvironment(accountName, envName string) (*Env, error)
	GetLoadMaps() (map[string]string, MountMap, error)

	ListMreses(envName string, options ...fn.Option) ([]Mres, error)
	ListMresKeys(envName, importedManagedResource string, options ...fn.Option) ([]string, error)
	GetMresConfigValues(options ...fn.Option) (map[string]string, error)

	ListSecrets(accountName string, envName string) ([]Secret, error)
	GetSecret(options ...fn.Option) (*Secret, error)
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
