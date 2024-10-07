package apiclient

import (
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type apiClient struct {
	fc fileclient.FileClient
}

type ApiClient interface {
	ListTeams() ([]Team, error)
	GetHostDNSSuffix() (string, error)

	ListApps(teamName string, envName string) ([]App, error)
	InterceptApp(app *App, status bool, ports []AppPort, envName string, options ...fn.Option) (err error)

	CreateRemoteLogin() (loginId string, err error)
	GetCurrentUser() (*User, error)
	Login(loginId string) error

	ListConfigs(teamName string, envName string) ([]Config, error)
	GetConfig(teamName string, envName string, configName string) (*Config, error)

	GetVPNDevice(teamName string, devName string) (*Device, error)
	CheckDeviceStatus() bool
	GetAccVPNConfig(team string) (*fileclient.TeamVpnConfig, error)
	CreateVpnForTeam(team string) (*Device, error)
	CreateDevice(devName, displayName, team string) (*Device, error)

	GetClusterConfig(team string) (*fileclient.TeamClusterConfig, error)

	ListEnvs(teamName string) ([]Env, error)
	GetEnvironment(teamName, envName string) (*Env, error)
	EnsureEnv() (*fileclient.Env, error)
	CloneEnv(teamName, envName, newEnvName, clusterName string) (*Env, error)
	CheckEnvName(teamName, envName string) (bool, error)
	GetLoadMaps() (map[string]string, MountMap, error)

	ListBYOKClusters(teamName string) ([]BYOKCluster, error)

	ListMreses(teamName string, envName string) ([]Mres, error)
	ListMresKeys(teamName, envName, importedManagedResource string) ([]string, error)
	GetMresConfigValues(teamName string) (map[string]string, error)

	ListSecrets(teamName string, envName string) ([]Secret, error)
	GetSecret(teamName string, secretName string) (*Secret, error)

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
