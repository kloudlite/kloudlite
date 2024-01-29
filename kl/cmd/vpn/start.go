package vpn

import (
	"os"
	"time"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/spf13/cobra"
)

// not required in linux

var connectVerbose bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start vpn device",
	Example: fn.Desc(`# start vpn device
sudo {cmd} vpn start`),
	Run: func(cmd *cobra.Command, _ []string) {
		if euid := os.Geteuid(); euid != 0 {
			fn.Log(
				text.Colored("make sure you are running command with sudo", 209),
			)
			return
		}

		options := []fn.Option{}

		if b := cmd.Flags().Changed("no-dns"); b {
			options = append(options, fn.MakeOption("noDns", "yes"))
		}

		switch flags.CliName {
		case constants.CoreCliName:
			envName := fn.ParseStringFlag(cmd, "env")
			projectName := fn.ParseStringFlag(cmd, "project")
			options = append(options, fn.MakeOption("projectName", projectName))
			options = append(options, fn.MakeOption("environmentName", envName))

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
	success := false

	defer func() {
		time.Sleep(200 * time.Millisecond)
		if !success {
			_ = wg_vpn.StopService(verbose)
		}
	}()

	configFolder, err := client.GetConfigFolder()
	if err != nil {
		return err
	}

	devName, err := client.CurrentDeviceName()
	if err != nil {
		return err
	}

	if err := wg_vpn.StartServiceInBg(devName, configFolder); err != nil {
		return err
	}

	if err := connect(verbose, options...); err != nil {
		return err
	}

	success = true
	return nil
}

func init() {
	startCmd.Flags().BoolVarP(&connectVerbose, "verbose", "v", false, "show verbose")
	startCmd.Flags().BoolP("no-dns", "n", false, "do not update dns")

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
