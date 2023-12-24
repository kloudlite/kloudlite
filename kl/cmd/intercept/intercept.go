package intercept

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "intercept",
	Short: "intercept an app with your device",
	Long: `Intercept app to tunnel your local
Examples:
  # intercept an app with device id and appid
  kl intercept <device_id> <app_id>

  # intercept an app with device name and app readable id
  kl intercept <device_name> <app_readable_id>

  # alternative way
  kl intercept --device-id=<divice_id> --app-readable-id=<app_readable_id>
	`,
	Run: func(cmd *cobra.Command, args []string) {
		err := interceptApp(cmd, args)
		if err != nil {
			common.PrintError(err)
			return
		}

		fmt.Println("Interception done")
	},
}

func interceptApp(cmd *cobra.Command, args []string) error {
	dName := cmd.Flag("device-name").Value.String()
	if dName == "" {
		dName = cmd.Flag("device-id").Value.String()
	}

	if len(args) >= 1 && dName == "" {
		dName = args[0]
	}

	appId, err := triggerSelectApp(cmd, args)
	if err != nil {
		return err
	}

	deviceId, err := triggerDeviceSelect(dName)
	if err != nil {
		return err
	}

	err = server.InterceptApp(deviceId, appId)

	return err
}

func triggerDeviceSelect(dName string) (string, error) {
	devices, err := server.GetDevices()
	if err != nil {
		return "", err
	}

	if dName != "" {
		for _, d := range devices {
			if d.Name == dName || d.Id == dName {
				return d.Id, nil
			}
		}
		return "", errors.New("provided device-name is not yours or not present in selected account")
	}

	selectedIndex, err := fuzzyfinder.Find(
		devices,
		func(i int) string {
			return devices[i].Name
		},
		fuzzyfinder.WithPromptString("Select Device >"),
	)

	if err != nil {
		return "", err
	}

	return devices[selectedIndex].Id, nil
}

func triggerSelectApp(cmd *cobra.Command, args []string) (string, error) {
	aId := cmd.Flag("app-id").Value.String()

	if aId == "" {
		aId = cmd.Flag("app-readable-id").Value.String()
	}

	if len(args) >= 2 && aId == "" {
		aId = args[1]
	}

	apps, err := server.GetApps()
	if err != nil {
		return "", err
	}

	if aId != "" {
		for _, a := range apps {
			if a.Id == aId || a.ReadableId == aId {
				return a.Id, nil
			}
		}
		return "", errors.New("provided app id not found in selected project")
	}

	selectedIndex, err := fuzzyfinder.Find(
		apps,
		func(i int) string {
			return apps[i].Name
		},
		fuzzyfinder.WithPromptString("Select App >"),
	)

	if err != nil {
		return "", err
	}

	return apps[selectedIndex].Id, nil
}

func init() {
	Cmd.Flags().StringP("device-id", "", "", "device id")
	Cmd.Flags().StringP("device-name", "", "", "device name")
	Cmd.Flags().StringP("app-id", "", "", "app/lambda id")
	Cmd.Flags().StringP("app-readable-id", "", "", "app/lambda  readable id")
}
