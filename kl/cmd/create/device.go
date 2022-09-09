package create

import (
	"errors"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "create new device",
	Long: `create device
Examples:
  # create new device
  kl create device <device_name>

	# Note: device name must not contain space or special character
	`,
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			common.PrintError(errors.New("device name not provided"))
			return
		}
		dName := args[0]
		err := server.CreateDevice(dName)
		if err != nil {
			common.PrintError(err)
			return
		}

		common.PrintError(errors.New("device created successfully"))
	},
}
