package device

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"

	"github.com/spf13/cobra"
)

var interceptCmd = &cobra.Command{
	Use:   "intercept",
	Short: "intercept an app with your device",
	Long: `Intercept app to tunnel your local
Examples:
  # intercept an app with device name and app name
  kl intercept <device_name> <app_name>

  # alternative way
  kl intercept --device-name=<device_name> --app-name=<app_name>
	`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := triggerSelectApp(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		fmt.Println("Interception done")
	},
}

func interceptApp(cmd *cobra.Command, args []string) error {
	// dName := cmd.Flag("device-name").Value.String()
	// if dName == "" {
	// 	dName = cmd.Flag("device-id").Value.String()
	// }
	//
	// if len(args) >= 1 && dName == "" {
	// 	dName = args[0]
	// }
	//
	// appId, err := triggerSelectApp(cmd, args)
	// if err != nil {
	// 	return err
	// }
	//
	// deviceId, err := triggerDeviceSelect(dName)
	// if err != nil {
	// 	return err
	// }
	//
	// err = server2.InterceptApp(deviceId, appId)
	//
	// return err

	return nil
}

func triggerDeviceSelect(dName string) (string, error) {
	// devices, err := server2.GetDevices()
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

func triggerSelectApp(cmd *cobra.Command, args []string) (string, error) {
	aName := cmd.Flag("app-name").Value.String()

	if len(args) >= 2 && aName == "" {
		aName = args[1]
	}
	//
	//if aName == "" {
	//	return "", errors.New("please provide an app name in the selected project")
	//}

	apps, err := server.ListApps()
	if err != nil {
		return "", err
	}

	if aName != "" {
		for _, a := range apps {
			if a.Metadata.Name == aName {
				return a.Metadata.Name, nil
			}
		}
		return "", errors.New("provided app name not found in selected project")
	}

	appName, err := fzf.FindOne(
		apps,
		func(item server.App) string {
			return item.Metadata.Name
		},
		fzf.WithPrompt("Select App >"),
	)

	if err != nil {
		return "", err
	}
	return appName.Metadata.Name, nil
}

func init() {
	Cmd.Flags().StringP("device-name", "", "", "device name")
	Cmd.Flags().StringP("app-name", "", "", "app/lambda name")
}
