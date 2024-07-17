package intercept

import (
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
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

		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		currentAcc, err := fc.CurrentAccountName()
		if err != nil {
			fn.PrintError(err)
			return
		}

		currentEnv, err := fc.CurrentEnv()
		if err != nil {
			fn.PrintError(err)
			return
		}

		apps, err := apic.ListApps(currentAcc, currentEnv.Name)
		if err != nil {
			fn.PrintError(err)
			return
		}

		filteredApps := make([]apiclient.App, 0)
		for _, app := range apps {
			if app.Spec.Intercept != nil && app.Spec.Intercept.Enabled {
				filteredApps = append(filteredApps, app)
			}
		}
		if len(filteredApps) == 0 {
			fn.Log("no intercepted apps found")
			return
		}

		appToStop, err := fzf.FindOne(filteredApps, func(item apiclient.App) string {
			return item.DisplayName
		}, fzf.WithPrompt("Select app to stop"))
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := apic.InterceptApp(appToStop, false, nil, currentEnv.Name, []fn.Option{
			fn.MakeOption("appName", appToStop.Metadata.Name),
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
