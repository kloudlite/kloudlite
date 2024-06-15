package flags

import (
	"os/exec"
)

var (
	Version    = "development"
	CliName    = "kl"
	BasePrefix = ""
	DevMode    = "false"
)

func GetCliPath() string {
	s, err := exec.LookPath(CliName)
	if err != nil {
		return CliName
	}
	return s
}

func IsDev() bool {
	if DevMode == "false" {
		return false
	}
	return true
}
