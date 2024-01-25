package kl

import (
	"fmt"

	"github.com/spf13/cobra"
)

func GetRootHelp(cmd *cobra.Command) string {
	return fmt.Sprintf(`Usage: kl [command] [options] [-- command] [command options]

These are common kl commands used in various situatiions:

Manage Context:
      status                      - get status of your current context (user, account, project, environment, vpn status)

      list account                - list all accounts
      switch account              - switch between kloudlite accounts

Setup a kloudlite environment:
      init                        - initilize kloudlite configuration file in current working directory
      add [config|secret|mres]    - add config,secret and managed resource entries to kloudlite configuration file

Working inside environment:
      intercept <appname>         - intercept the application in the environment with your device.
                                    This will tunnel all the incoming traffic to your device

      -- <command>                - execute any command with loaded env variables
                                    Example: kl -- npm start

      list env                    - list all environments in current project
      switch environment          - inside the current project context you can switch between environments

VPN management:
      vpn connect                 - connect/switch your device to current working environment (requires sudo)
      vpn disconnect              - disconnect your device (requires sudo)
      vpn [wg | status]           - wg will show the (handshake, peers, etc) and status will show the connection status
      vpn expose                  - expose your local device ports

Fetch resources of current environment:
      get [config | secret | mres] <name>    - get config entries
      list [configs | secrets | mres]        - list all configs,secrets,mres in current environment of project

Other Available Commands:
      auth                        - login, logout, whoami

	`)
}
