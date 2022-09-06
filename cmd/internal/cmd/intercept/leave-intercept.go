package intercept

import (
	"errors"
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/lib/server"
)

var leaveInterceptCmd = &cobra.Command{
	Use:   "intercept",
	Short: "close interception of an app which is intercepted",
	Long: `Close Interception
Examples:
	# close interception of app with device id and appid
	kl leave intercept <device_id> <app_id>

	# close interception an app with device name and app readable id
	kl leave intercept <device_name> <app_readable_id>

	# alternative way
	kl leave intercept --device-id=<divice_id> --app-readable-id=<app_readable_id>
`,

	Run: func(cmd *cobra.Command, _ []string) {
		err := closeInterceptApp(cmd)
		if err != nil {
			common.PrintError(err)
			return
		}

		fmt.Println("Interception closed")
	},
}

func closeInterceptApp(cmd *cobra.Command) error {

	dName := cmd.Flag("device-name").Value.String()
	dId := cmd.Flag("device-id").Value.String()

	device, err := deviceSelect(dId, dName)
	if err != nil {
		return err
	}

	appId, err := selectApp(*device, cmd)
	if err != nil {
		return err
	}

	err = server.CloseInterceptApp(appId)
	return err
}

func selectApp(device server.Device, cmd *cobra.Command) (string, error) {
	arId := cmd.Flag("app-readable-id").Value.String()
	aId := cmd.Flag("app-id").Value.String()

	apps := device.Intercepted
	if len(apps) == 0 {
		return "", errors.New("not apps intercepted by this device")
	}

	if arId != "" {
		for _, d := range apps {
			if d.Id == arId {
				return d.Id, nil
			}
		}

		return "", errors.New("provided app is not intercepted")
	}

	if aId != "" {
		for _, d := range apps {
			if d.ReadableId == arId {
				return d.Id, nil
			}
		}

		return "", errors.New("provided app is not intercepted")
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

func deviceSelect(dId, dName string) (*server.Device, error) {
	d, err := server.GetDevices()
	if err != nil {
		return nil, err
	}

	devices := []server.Device{}
	for _, d2 := range d {
		if len(d2.Intercepted) > 0 {
			devices = append(devices, d2)
		}
	}

	if len(devices) == 0 {
		return nil, errors.New("there is no apps intercepted by you")
	}

	if dId != "" {
		for _, d2 := range devices {
			if dId != d2.Id {
				return &d2, nil
			}
		}
		return nil, errors.New("no apps Intercepted by provided device")
	}

	if dName != "" {
		for _, d2 := range devices {
			if dName != d2.Name {
				return &d2, nil
			}
		}
		return nil, errors.New("no apps Intercepted by provided device")
	}

	selectedIndex, err := fuzzyfinder.Find(
		devices,
		func(i int) string {
			return fmt.Sprintf("%s | %d app%s intercepted", devices[i].Name, len(devices[i].Intercepted), func() string {
				if len(devices[i].Intercepted) <= 1 {
					return ""
				}
				return "s"
			}())
		},
		fuzzyfinder.WithPromptString("Select Device >"),
	)

	if err != nil {
		return nil, err
	}

	return &devices[selectedIndex], nil
}

func init() {
	leaveInterceptCmd.Flags().StringP("device-id", "", "", "device id")
	leaveInterceptCmd.Flags().StringP("device-name", "", "", "device name")
	leaveInterceptCmd.Flags().StringP("app-id", "", "", "app/lambda id")
	leaveInterceptCmd.Flags().StringP("app-readable-id", "", "", "app/lambda  readable id")

}
