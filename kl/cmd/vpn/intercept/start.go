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

  # intercept app with specified vpn device
  kl vpn intercept start --app <app_name> --device <device_name>
	`,

	Run: func(cmd *cobra.Command, _ []string) {
		app := fn.ParseStringFlag(cmd, "app")
		device := fn.ParseStringFlag(cmd, "device")
		env := fn.ParseStringFlag(cmd, "env")
		project := fn.ParseStringFlag(cmd, "project")

		err := server.InterceptApp(true, []fn.Option{
			fn.MakeOption("appName", app),
			fn.MakeOption("deviceName", device),
			fn.MakeOption("envName", env),
			fn.MakeOption("projectName", project),
		}...)

		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("intercept app started successfully")
	},
}

func init() {
	startCmd.Flags().StringP("app", "a", "", "app name")
	startCmd.Flags().StringP("device", "d", "", "device name")
	startCmd.Flags().StringP("env", "e", "", "environment name")
	startCmd.Flags().StringP("project", "p", "", "project name")
}
