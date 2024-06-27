package intercept

import (
	"github.com/kloudlite/kl/domain/apiclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [app_name]",
	Short: "stop tunneling the traffic to your device",
	Long: `stop intercept app to stop tunnel traffic to your device
Examples:
	# close intercept app
  kl intercept stop [app_name]
	`,

	Run: func(cmd *cobra.Command, _ []string) {
		ns := ""

		// app := fn.ParseStringFlag(cmd, "app")
		app := ""

		if cmd.Flags().Changed("name") {
			ns, _ = cmd.Flags().GetString("name")
		}

		if ns == "" {
			e, err := apiclient.EnsureEnv(nil)
			if err != nil {
				fn.PrintError(err)
				return
			}

			ns = e.TargetNs
		}

		if err := apiclient.InterceptApp(false, nil, []fn.Option{
			fn.MakeOption("appName", app),
		}...); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("intercepted app stopped successfully")
	},
}

func init() {
	// stopCmd.Flags().StringP("app", "a", "", "app name")

	stopCmd.Aliases = append(startCmd.Aliases, "close", "end", "leave", "quit", "terminate", "exit", "remove", "disconnect")
}
