package device

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "device",
	Short: "device specific commands",
	Long: `Device specific commands
Examples:
  # expose port of selected device
  kl device expose port -p <port>:<your_local_port>

  # intercept an app with device name and app name
  kl device intercept <device_name> <app_name>

  # alternative way
  kl device intercept --device-name=<device_name> --app-name=<app_name>

  # close interception of app with device id and appname
  kl device leave intercept <device_id> <app_name>
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "dev")

	Cmd.AddCommand(interceptCmd)
}
