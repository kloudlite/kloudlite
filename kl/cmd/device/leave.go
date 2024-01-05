package device

import (
	"github.com/spf13/cobra"
)

var LeaveCmd = &cobra.Command{
	Use:   "leave",
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
}

func init() {
	LeaveCmd.AddCommand(leaveInterceptCmd)
}
