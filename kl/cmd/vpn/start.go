package vpn

import (
	"os"
	"runtime"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/spf13/cobra"
)

// not required in linux

var connectVerbose bool

var skipCheck bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start vpn device",
	Example: fn.Desc(`# start vpn device
sudo {cmd} vpn start`),
	Run: func(cmd *cobra.Command, _ []string) {

		if runtime.GOOS == constants.RuntimeWindows {
			if err := connect(connectVerbose); err != nil {
				fn.Notify("Error:", err.Error())
				fn.PrintError(err)
			}
			return
		}

		if euid := os.Geteuid(); euid != 0 {
			fn.Log(
				text.Colored("make sure you are running command with sudo", 209),
			)
			return
		}

		options := []fn.Option{}

		switch flags.CliName {
		case constants.CoreCliName:
			envName := fn.ParseStringFlag(cmd, "env")
			if envName == "" {
				klFile, err := client.GetKlFile("")
				if err != nil && !os.IsNotExist(err) {
					fn.PrintError(err)
					return
				}
				if !os.IsNotExist(err) {
					envName = klFile.DefaultEnv
				}
			}
			projectName := fn.ParseStringFlag(cmd, "project")
			options = append(options, fn.MakeOption("projectName", projectName))
			options = append(options, fn.MakeOption("envName", envName))

		case constants.InfraCliName:
			clusterName := fn.ParseStringFlag(cmd, "cluster")
			options = append(options, fn.MakeOption("clusterName", clusterName))
		}

		wgInterface, err := wgc.Show(&wgc.WgShowOptions{
			Interface: "interfaces",
		})

		if err != nil {
			fn.PrintError(err)
			return
		}

		if len(wgInterface) != 0 {
			fn.Log("[#] already connected")

			fn.Log("\n[#] reconnecting")

			if err := disconnect(connectVerbose); err != nil {
				fn.PrintError(err)
				return
			}

			if err := startConnecting(connectVerbose, options...); err != nil {
				fn.PrintError(err)
				return
			}

			fn.Log("[#] connected")
			fn.Log("[#] reconnection done")

			return
		}

		if err := startConnecting(connectVerbose, options...); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("[#] connected")

		_, err = wgc.Show(nil)

		if err != nil {
			fn.PrintError(err)
			return
		}

		s, err := client.CurrentDeviceName()
		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(text.Bold(text.Green("\n[#]Selected Device:")),
			text.Red(s),
		)
	},
}

func startConnecting(verbose bool, options ...fn.Option) error {

	if err := connect(verbose, options...); err != nil {
		if skipCheck {
			fn.Notify("Error: ", err.Error())
		}
		return err
	}

	return nil
}

func init() {
	startCmd.Flags().BoolVarP(&connectVerbose, "verbose", "v", false, "show verbose")
	startCmd.Flags().BoolVarP(&skipCheck, "skipCheck", "s", false, "skip checks of env and cluster")
	startCmd.Aliases = append(stopCmd.Aliases, "connect")

	switch flags.CliName {
	case constants.CoreCliName:
		{
			startCmd.Flags().StringP("project", "p", "", "project name")
			startCmd.Flags().StringP("env", "e", "", "environment name")
		}
	case constants.InfraCliName:
		{
			startCmd.Flags().StringP("cluster", "c", "", "cluster name")
		}
	}

}
