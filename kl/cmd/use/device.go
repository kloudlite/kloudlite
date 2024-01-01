package use

import (
	common_util "github.com/kloudlite/kl/pkg/functions"

	"github.com/kloudlite/kl/lib"
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

		deviceId, err := triggerDeviceSelect(dName)
		if err != nil {
			common_util.PrintError(err)
			return
		}

		err = lib.SelectDevice(deviceId)
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}

func triggerDeviceSelect(dName string) (string, error) {
	// devices, err := server.GetDevices()
	// if err != nil {
	// 	return "", err
	// }
	//
	// if dName != "" {
	// 	for _, d := range devices {
	// 		if d.Name == dName || d.Id == dName {
	// 			return d.Id, nil
	// 		}
	// 	}
	// 	return "", errors.New("provided device-name is not yours or not present in selected account")
	// }
	//
	// selectedIndex, err := fuzzyfinder.Find(
	// 	devices,
	// 	func(i int) string {
	// 		return devices[i].Name
	// 	},
	// 	fuzzyfinder.WithPromptString("Select Device >"),
	// )
	//
	// if err != nil {
	// 	return "", err
	// }
	//
	// return devices[selectedIndex].Id, nil

	return "", nil
}
