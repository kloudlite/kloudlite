package vpn

import (
	"fmt"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "listing all contexts",
	Long: `This command let you list all contexts.
Example:
  # list all contexts
  kl context list
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
		return fmt.Errorf("no vpn devices found")
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
	listCmd.Aliases = append(listCmd.Aliases, "ls")
}
