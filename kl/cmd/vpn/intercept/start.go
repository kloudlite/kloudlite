package intercept

import (
	"strconv"
	"strings"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [app_name]",
	Short: "start tunneling the traffic to your device",
	Long: `start intercept app to tunnel trafic to your device
Examples:
	# intercept app with selected vpn device
  kl intercept start [app_name] --port <port>:<your_local_port>
	`,

	Run: func(cmd *cobra.Command, _ []string) {
		// app := fn.ParseStringFlag(cmd, "app")
		app := ""
		maps, err := cmd.Flags().GetStringArray("port")
		if err != nil {
			fn.PrintError(err)
			return
		}

		ports := make([]apiclient.AppPort, 0)

		for _, v := range maps {
			mp := strings.Split(v, ":")
			if len(mp) != 2 {
				fn.PrintError(
					functions.Error("wrong map format use <server_port>:<local_port> eg: 80:3000"),
				)
				return
			}

			pp, err := strconv.ParseInt(mp[0], 10, 32)
			if err != nil {
				fn.PrintError(err)
				return
			}

			tp, err := strconv.ParseInt(mp[1], 10, 32)
			if err != nil {
				fn.PrintError(err)
				return
			}

			ports = append(ports, apiclient.AppPort{
				AppPort:    int(pp),
				DevicePort: int(tp),
			})
		}

		err = apiclient.InterceptApp(true, ports, []fn.Option{
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
	// startCmd.Flags().StringP("app", "a", "", "app name")
	startCmd.Flags().StringArrayP(
		"port", "p", []string{},
		"expose port <server_port>:<local_port> while intercepting app",
	)

	startCmd.Aliases = append(startCmd.Aliases, "add", "begin", "connect")
}
