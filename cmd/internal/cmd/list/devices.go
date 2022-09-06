package list

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/common/ui/table"
	"kloudlite.io/cmd/internal/lib/server"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "list all your devices in selected account",
	Long: `List all your devices in selected account.
Examples:
	# list all the devices with selected account
  kl list devices

	# list all the devices with accountId
  kl list devices <accountId>
`,
	Run: func(_ *cobra.Command, args []string) {
		err := listDevices(args)
		if err != nil {
			common.PrintError(err)
			return
		}
	},
}

func listDevices(args []string) error {
	var devices []server.Device
	var err error
	if len(args) >= 1 {
		devices, err = server.GetDevices(common.MakeOption("accountId", ""))
	} else {
		devices, err = server.GetDevices()
	}

	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return errors.New("no Devices found")
	}

	header := table.Row{
		table.HeaderText("name, id"),
		table.HeaderText("Active Region"),
		table.HeaderText("exposed ports"),
	}

	rows := make([]table.Row, 0)

	for _, a := range devices {
		rows = append(rows, table.Row{fmt.Sprintf("%s, %s", a.Name, a.Id),
			a.Region,
			strings.Join(func() []string {
				ports := []string{}

				for i, v := range a.Ports {
					ports = append(ports, fmt.Sprintf("%s%d:%d", func() string {
						if (i+1)%3 == 0 {
							return "\n"
						}
						return ""
					}(), v.Port, func() int {
						if v.TargetPort == 0 {
							return v.Port

						}
						return v.TargetPort
					}()))
				}

				return ports
			}(), ", ")})
	}

	fmt.Println(table.Table(header, rows))

	if accountId, _ := server.CurrentAccountId(); accountId != "" {
		fmt.Println(table.KVOutput("devices of", accountId))
	}

	fmt.Println(table.TotalResults(len(devices)))
	return nil
}

func init() {
	devicesCmd.Aliases = append(devicesCmd.Aliases, "device")
}
