package constants

import (
	"fmt"

	"github.com/kloudlite/kl/flags"

	"github.com/kloudlite/kl/domain/fileclient"
)

const (
	RuntimeLinux                = "linux"
	RuntimeDarwin               = "darwin"
	RuntimeWindows              = "windows"
	KLDNS                       = "100.64.0.1"
	InterceptWorkspaceServiceIp = "172.18.0.3"
	K3sServerIp                 = "172.18.0.2"
	//SocatImage                  = "ghcr.io/kloudlite/hub/socat:latest"
)

var DefaultBaseURL = flags.DefaultBaseURL

func GetWireguardImageName() string {
	return fmt.Sprintf("%s/box/wireguard:%s", flags.ImageBase, flags.Version)
}

func GetK3SImageName() string {
	return fmt.Sprintf("%s/k3s:%s", flags.ImageBase, flags.Version)
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
