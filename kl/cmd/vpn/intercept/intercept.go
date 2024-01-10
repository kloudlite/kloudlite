package intercept

import (
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "intercept",
	Short: "intercept app to tunnel trafic to your device",
	Long: `intercept app to tunnel trafic to your device
Examples:
	# intercept app with selected vpn device
  kl vpn intercept start --app <app_name>

	# stop intercept app with selected vpn device
	kl vpn intercept stop --app <app_name>

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
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.Aliases = append(startCmd.Aliases, "inc")

	Cmd.Flags().StringP("app", "a", "", "app name")
	Cmd.Flags().StringP("device", "d", "", "device name")
	Cmd.Flags().StringP("env", "e", "", "environment name")
	Cmd.Flags().StringP("project", "p", "", "project name")
}
