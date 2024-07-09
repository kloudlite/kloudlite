package intercept

import (
	"slices"
	"strconv"
	"strings"

	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/envclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"

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
		if err := startIntercept(apic, fc, cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func startIntercept(apic apiclient.ApiClient, fc fileclient.FileClient, cmd *cobra.Command, args []string) error {
	accName, err := fc.CurrentAccountName()
	if err != nil {
		return err
	}
	currentEnv, err := fc.CurrentEnv()
	if err != nil {
		return err
	}

	appsList, err := apic.ListApps(accName, currentEnv.Name)
	if err != nil {
		return err
	}

	selectedApp, err := fzf.FindOne[apiclient.App](appsList, func(item apiclient.App) string {
		return item.DisplayName
	}, fzf.WithPrompt("Select app to intercept"))
	if err != nil {
		return err
	}

	// app := ""
	maps, err := cmd.Flags().GetStringArray("port")
	if err != nil {
		return err
	}

	ports := make([]apiclient.AppPort, 0)

	for _, v := range maps {
		mp := strings.Split(v, ":")
		if len(mp) != 2 {
			return fn.Error("wrong map format use <server_port>:<local_port> eg: 80:3000")
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

	err = apic.InterceptApp(selectedApp, true, ports, currentEnv.Name, []fn.Option{
		fn.MakeOption("appName", selectedApp.Metadata.Name),
	}...)

	if err != nil {
		return err
	}

	bc, err := boxpkg.NewClient(cmd, args)
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
	eports = append(eports, kt.Ports...)

	for _, v := range ports {
		if !slices.Contains(eports, v.DevicePort) {
			eports = append(eports, v.DevicePort)
		}
	}

	if err := bc.SyncProxy(boxpkg.ProxyConfig{
		TargetContainerPath: s,
		ExposedPorts:        eports,
	}); err != nil {
		return err
	}

	p := kt.Ports
	for _, v := range ports {
		if !slices.Contains(p, v.DevicePort) {
			p = append(p, v.DevicePort)
		}
	}
	kt.Ports = p
	if err = fc.WriteKLFile(*kt); err != nil {
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
