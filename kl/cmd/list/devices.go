package list

import (
	"fmt"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "list all your devices in selected account",
	Long: `List all your devices in selected account.

This command will provide the list of all the devices in the selected account.
Examples:
  kl list [devices|device|dev|devs]

Note: selected project will be highlighted with green color.

`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := listDevices(); err != nil {
			fn.PrintError(err)
		}
	},
}

func listDevices() error {

	devices, err := server.ListDevices()
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return fmt.Errorf("no devices found")
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		table.HeaderText("Active_Ns"),
		table.HeaderText("Ports"),
	}

	rows := make([]table.Row, 0)
	activeDevName, _ := client.CurrentDeviceName()

	for _, d := range devices {
		rows = append(rows, table.Row{
			fn.GetPrintRow(d, activeDevName, d.DisplayName, true),
			fn.GetPrintRow(d, activeDevName, d.Metadata.Name),
			fn.GetPrintRow(d, activeDevName, d.Spec.DeviceNamespace),
			fn.GetPrintRow(d, activeDevName, func() string {
				if d.Spec.Ports == nil {
					return ""
				}

				res := make([]string, 0)

				for _, p := range d.Spec.Ports {
					res = append(res, fmt.Sprintf("%d:%d ", p.Port, func() int {
						if p.TargetPort == 0 {
							return p.Port
						}
						return p.TargetPort
					}()))
				}

				return strings.Join(res, "\n")
			}()),
		})
	}

	fmt.Println(table.Table(&header, rows))
	table.TotalResults(len(devices), true)

	return nil
}

func init() {
	devicesCmd.Aliases = append(devicesCmd.Aliases, "device")
	devicesCmd.Aliases = append(devicesCmd.Aliases, "dev")
	devicesCmd.Aliases = append(devicesCmd.Aliases, "devs")
}
