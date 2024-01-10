package vpn

import (
	"fmt"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "switch between different vpn devices",
	Long: `This command let switch between different vpn devices.
Example:
  # switch to vpn device
  kl vpn switch

	# switch to vpn device with device name
	kl vpn switch --name <device_name>
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		name := ""
		if cmd.Flags().Changed("name") {
			name, _ = cmd.Flags().GetString("name")
		}

		d, err := server.SelectDevice(name)
		if err != nil {
			fn.PrintError(err)
			return
		}

		fmt.Println("Selected vpn device: ", d.Metadata.Name)

	},
}

func init() {
	switchCmd.Aliases = append(switchCmd.Aliases, "sw")
	switchCmd.Flags().StringP("name", "n", "", "device name")
}
