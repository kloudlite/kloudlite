package constants

import (
	"fmt"

	"github.com/kloudlite/kl/flags"

	"github.com/kloudlite/kl/domain/fileclient"
)

const (
	DefaultBaseURL              = "https://auth.kloudlite.io"
	RuntimeLinux                = "linux"
	RuntimeDarwin               = "darwin"
	RuntimeWindows              = "windows"
	SocatImage                  = "ghcr.io/kloudlite/hub/socat:latest"
	KLDNS                       = "100.64.0.1"
	InterceptWorkspaceServiceIp = "172.18.0.3"
	K3sServerIp                 = "172.18.0.2"
)

func GetWireguardImageName() string {
	return fmt.Sprintf("ghcr.io/kloudlite/kl/box/wireguard:%s", flags.Version)
}

func GetK3SImageName() string {
	//return "rancher/k3s:v1.27.5-k3s1"
	return fmt.Sprintf("ghcr.io/kloudlite/kl/k3s:%s", flags.Version)
}

var (
	BaseURL = func() string {
		baseUrl := DefaultBaseURL

		s, err := fileclient.GetBaseURL()
		if err == nil && s != "" {
			baseUrl = s
		}

		return baseUrl
	}()

	LoginUrl = func() string {
		return fmt.Sprintf("%s/cli-login", BaseURL)
	}()
	ServerURL = func() string {
		return fmt.Sprintf("%s/api/", BaseURL)
	}()

	UpdateURL = func() string {
		return "https://kl.kloudlite.io/kloudlite"
	}()
)

var (
	CoreCliName = "kl"
)
