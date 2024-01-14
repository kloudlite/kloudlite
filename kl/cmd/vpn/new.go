package vpn

import (
	"fmt"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/input"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new context",
	Long: `This command let create new context.
Example:
  # create new context
  kl vpn new

	# creating new context with name
	kl vpn new --device <device_name>
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		deviceName := ""
		if cmd.Flags().Changed("device") {
			deviceName, _ = cmd.Flags().GetString("device")
		}
		if deviceName == "" {
			var err error
			deviceName, err = input.Prompt(input.Options{
				Placeholder: "Enter device name",
				CharLimit:   25,
				Password:    false,
			})
			if err != nil {
				fn.PrintError(err)
				return
			}
		}
		if deviceName == "" {
			fn.PrintError(fmt.Errorf("device name is required"))
			return
		}
		suggestedNames, err := server.GetDeviceName(deviceName)
		if err != nil {
			fn.PrintError(err)
			return
		}
		selectedDeviceName := ""
		if suggestedNames.Result == true {
			selectedDeviceName = deviceName
		} else {
			selectedDeviceName, err = server.SelectDeviceName(suggestedNames.SuggestedNames)
			if err != nil {
				fn.PrintError(err)
				return
			}
		}
		device, err := server.CreateDevice(selectedDeviceName, deviceName)
		//device, err := server.CreateDevice(deviceName, deviceName)
		if err != nil {
			fn.PrintError(err)
			return
		}
		fn.Log(fmt.Sprintf("device %s has been created\n", device.Metadata.Name))
	},
}

func init() {
	newCmd.Flags().StringP("device", "d", "", "device name")
}
