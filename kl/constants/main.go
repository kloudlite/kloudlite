package constants

import (
	"fmt"
	"github.com/kloudlite/kl/flags"

	"github.com/kloudlite/kl/domain/fileclient"
)

const (
	DefaultBaseURL = "https://auth.kloudlite.io"
	RuntimeLinux   = "linux"
	RuntimeDarwin  = "darwin"
	RuntimeWindows = "windows"
	SocatImage = "ghcr.io/kloudlite/hub/socat:latest"
)

func GetWireguardImageName() string {
	return fmt.Sprintf("ghcr.io/kloudlite/kl/box/wireguard:%s", flags.Version)
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
