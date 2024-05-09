package intercept

import (
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

	//Run: func(cmd *cobra.Command, _ []string) {
	//	app := fn.ParseStringFlag(cmd, "app")
	//
	//	err := server.InterceptApp(true, []fn.Option{
	//		fn.MakeOption("appName", app),
	//	}...)
	//
	//	if err != nil {
	//		fn.PrintError(err)
	//		return
	//	}
	//
	//	fn.Log("intercept app started successfully")
	//},
}

func init() {
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.Aliases = append(startCmd.Aliases, "inc")

	Cmd.Flags().StringP("app", "a", "", "app name")
}
