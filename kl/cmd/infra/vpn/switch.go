package vpn

import (
	"github.com/kloudlite/kl/domain/client"

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
  kl infra vpn switch

	# switch to vpn device with device name
	kl infra vpn switch
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		name := ""
		if cmd.Flags().Changed("name") {
			name, _ = cmd.Flags().GetString("name")
		}

		d, err := server.SelectInfraDevice(name)
		if err != nil {
			fn.PrintError(err)
			return
		}

		activeInfraContext, err := client.GetActiveInfraContext()
		if err != nil {
			fn.PrintError(err)
			return
		}
		activeInfraContext.DeviceName = d.Metadata.Name

		err = client.WriteInfraContextFile(*activeInfraContext)
		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("Selected vpn device: ", d.Metadata.Name)

	},
}

func init() {
	switchCmd.Aliases = append(switchCmd.Aliases, "sw")
	switchCmd.Flags().StringP("name", "n", "", "device name")
}
