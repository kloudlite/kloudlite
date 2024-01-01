package list

import (
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "list all your devices in selected account",
	Long: `List all your devices in selected account.
Examples:
	# list all the devices with selected account
  kl list [devices|device|dev|devs]
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
	}

	rows := make([]table.Row, 0)
	activeDevName, _ := client.CurrentDeviceName()

	for _, d := range devices {

		rows = append(rows, table.Row{
			func() string {
				if d.Metadata.Name == activeDevName {
					return text.Colored(fmt.Sprintf("*%s", d.DisplayName), 2)
				}
				return d.DisplayName
			}(),

			func() string {
				if d.Metadata.Name == activeDevName {
					return text.Colored(fmt.Sprintf("*%s", d.Metadata.Name), 2)
				}
				return d.Metadata.Name
			}(),
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
