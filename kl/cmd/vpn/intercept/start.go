package intercept

import (
	"slices"
	"strconv"
	"strings"

	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/envclient"
	"github.com/kloudlite/kl/domain/fileclient"
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

	Run: func(cmd *cobra.Command, args []string) {
		if err := startIntercept(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func startIntercept(cmd *cobra.Command, args []string) error {
	// app := fn.ParseStringFlag(cmd, "app")
	app := ""
	maps, err := cmd.Flags().GetStringArray("port")
	if err != nil {
		return err
	}

	ports := make([]apiclient.AppPort, 0)

	for _, v := range maps {
		mp := strings.Split(v, ":")
		if len(mp) != 2 {
			return functions.Error("wrong map format use <server_port>:<local_port> eg: 80:3000")
		}

		pp, err := strconv.ParseInt(mp[0], 10, 32)
		if err != nil {
			return err
		}

		tp, err := strconv.ParseInt(mp[1], 10, 32)
		if err != nil {
			return err
		}

		ports = append(ports, apiclient.AppPort{
			AppPort:    int(pp),
			DevicePort: int(tp),
		})
	}

	err = apiclient.InterceptApp(true, ports, []fn.Option{
		fn.MakeOption("appName", app),
	}...)

	bc, err := boxpkg.NewClient(cmd, args)
	if err != nil {
		return err
	}

	fc, err := fileclient.New()
	if err != nil {
		return err
	}

	kt, err := fc.GetKlFile("")
	if err != nil {
		return err
	}

	s, err := envclient.GetWorkspacePath()
	if err != nil {
		return err
	}

	eports := []int{}
	for _, v := range kt.Ports {
		eports = append(eports, v)
	}

	for _, v := range ports {
		if !slices.Contains(eports, v.DevicePort) {
			eports = append(eports, v.DevicePort)
		}
	}

	bc.SyncProxy(boxpkg.ProxyConfig{
		TargetContainerPath: s,
		ExposedPorts:        eports,
	})

	if err != nil {
		return err
	}

	fn.Log("intercept app started successfully\n")
	fn.Log("Please check if vpn is connected to your device, if not please connect it using sudo kl vpn start. Ignore this message if already connected.")

	return nil
}

func init() {
	// startCmd.Flags().StringP("app", "a", "", "app name")
	startCmd.Flags().StringArrayP(
		"port", "p", []string{},
		"expose port <server_port>:<local_port> while intercepting app",
	)

	startCmd.Aliases = append(startCmd.Aliases, "add", "begin", "connect")
}
