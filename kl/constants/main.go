package constants

import (
	"fmt"

	"github.com/kloudlite/kl/domain/client"
)

const (
	DefaultBaseURL              = "https://auth.kloudlite.io"
	RuntimeLinux                = "linux"
	RuntimeDarwin               = "darwin"
	RuntimeWindows              = "windows"
	BashShell                   = "bash"
	FishShell                   = "fish"
	ZshShell                    = "zsh"
	PowerShell                  = "powershell"
	NetworkService              = "Wi-Fi"
	LocalSearchDomains          = ".local"
	NoExistingSearchDomainError = "There aren't any Search Domains set on Wi-Fi."
)

var (
	BaseURL = func() string {
		baseUrl := DefaultBaseURL

		s, err := client.GetBaseURL()
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
		return "https://kl.kloudlite.io/kloudlite/kl"
	}()
)

var (
	InfraCliName = "kli"
	CoreCliName  = "kl"
)
