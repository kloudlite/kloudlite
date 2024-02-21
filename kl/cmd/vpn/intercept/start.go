package intercept

import (
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start intercept app to tunnel trafic to your device",
	Long: `start intercept app to tunnel trafic to your device
Examples:
	# intercept app with selected vpn device
  kl vpn intercept start --app <app_name>
	`,

	Run: func(cmd *cobra.Command, _ []string) {
		app := fn.ParseStringFlag(cmd, "app")

		err := server.InterceptApp(true, []fn.Option{
			fn.MakeOption("appName", app),
		}...)

		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("intercept app started successfully\n")
		fn.Log("Please check if vpn is connected to your device, if not please connect it using sudo kl vpn start. Ignore this message if already connected.")
	},
}

func init() {
	startCmd.Flags().StringP("app", "a", "", "app name")

	startCmd.Aliases = append(startCmd.Aliases, "add", "begin", "connect")
}
