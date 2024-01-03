package use

import (
	"fmt"

	"github.com/kloudlite/kl/domain/server"
	common_util "github.com/kloudlite/kl/pkg/functions"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "select device to use later to intercept app or wireguard connect",
	Long: `Select device
Examples:
  # select device
  kl use device

	# this will open selector where you can select one of your device to use later to intercept app or wireguard connect.

  # select device with device ip
  kl use device <deviceId>

  # select device with device name
  kl use device <device_name>
`,
	Run: func(_ *cobra.Command, args []string) {

		dName := ""

		if len(args) >= 1 {
			dName = args[0]
		}

		d, err := server.SelectDevice(dName)
		if err != nil {
			common_util.PrintError(err)
			return
		}

		fmt.Println("Selected device: ", d.Metadata.Name)
	},
}

func init() {
	deviceCmd.Aliases = append(deviceCmd.Aliases, "dev")
}
