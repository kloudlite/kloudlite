package boxpkg

import (
	"fmt"

	"github.com/kloudlite/kl/flags"
)

const (
	// CONTAINER_PORT = "1729"
	VpnImageName = "linuxserver/wireguard:latest"

	CONT_PATH_KEY           = "kl.container.path"
	CONT_NAME_KEY           = "kl.container.name"
	CONT_MARK_KEY           = "kl.container"
	CONT_VPN_MARK_KEY       = "kl.container.vpn"
	CONT_WORKSPACE_MARK_KEY = "kl.container.workspace"
	SSH_PORT_KEY            = "kl.container.ssh.port"
	KLCONFIG_HASH_KEY       = "kl.container.klconfig.hash"
)

func GetImageName() string {
	return fmt.Sprintf("%s/box:%s", flags.ImageBase, flags.Version)
}
