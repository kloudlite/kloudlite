package vpn

import (
	"os"
	"runtime"
	"time"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/spf13/cobra"
)

// not required in linux

var skipCheck bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start vpn device",
	Example: fn.Desc(`# start vpn device
sudo {cmd} vpn start`),
	Run: func(cmd *cobra.Command, _ []string) {

		verbose := fn.ParseBoolFlag(cmd, "verbose")

		if os.Getenv("KL_APP") != "true" {
			if euid := os.Geteuid(); euid != 0 {

				if err := func() error {

					if err := client.EnsureAppRunning(); err != nil {
						return err
					}

					p, err := proxy.NewProxy(true)
					if err != nil {
						return err
					}

					if err := p.Start(); err != nil {
						return err
					}

					return nil
				}(); err != nil {
					fn.PrintError(err)
					return
				}

				return
			}
		}

		options := []fn.Option{}

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

			if err := disconnect(verbose); err != nil {
				fn.PrintError(err)
				return
			}

			if runtime.GOOS == constants.RuntimeWindows {
				time.Sleep(3 * time.Second)
			}

			time.Sleep(1 * time.Second)

			if err := startConnecting(verbose, options...); err != nil {
				fn.PrintError(err)
				return
			}

		} else {
			if err := startConnecting(verbose, options...); err != nil {
				fn.PrintError(err)
				return
			}
		}

		s, err := client.CurrentDeviceName()
		if err != nil {
			fn.Logf(text.Bold("\n[#] connection to device done"))
			fn.PrintError(err)
			return
		}

		fn.Logf(text.Bold("\n[#] connection to device %s%s\n"), text.Blue(s), text.Bold(" done"))
	},
}

func startConnecting(verbose bool, options ...fn.Option) error {
	if err := connect(verbose, options...); err != nil {
		return err
	}

	return nil
}

func init() {
	startCmd.Flags().BoolP("verbose", "v", false, "run in debug mode")
	startCmd.Flags().BoolVarP(&skipCheck, "skipCheck", "s", false, "skip checks of env and cluster")
	startCmd.Aliases = append(stopCmd.Aliases, "connect")
}
