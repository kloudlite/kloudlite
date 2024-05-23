package intercept

import (
	"github.com/kloudlite/kl/domain/client"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "intercept",
	Short: "intercept app to tunnel trafic to your device",
	Long: `intercept app to tunnel trafic to your device
Examples:
	# intercept app with selected vpn device
  kl vpn intercept start --app <app_name> --port <port>:<your_local_port> 

	# stop intercept app with selected vpn device
	kl vpn intercept stop --app <app_name>
	`,
}

func init() {

	client.OnlyInsideBox(startCmd)
	Cmd.AddCommand(startCmd)

	client.OnlyInsideBox(stopCmd)
	Cmd.AddCommand(stopCmd)

	Cmd.Aliases = append(startCmd.Aliases, "inc")
	Cmd.Flags().StringP("app", "a", "", "app name")
}
