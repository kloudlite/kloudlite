package account

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
	"os"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch account",
	Long: `Use this command to switch account
Example:
  # switch account context
  kl account switch
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		accountName := ""

		accountName = fn.ParseStringFlag(cmd, "account")

		a, err := server.SelectAccount(accountName)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.WriteAccountContext(a.Metadata.Name); err != nil {
			fn.PrintError(err)
			return
		}

		if err != nil {
			fn.PrintError(err)
			return
		}
		devName, err := os.Hostname()
		if err != nil {
			fn.PrintError(err)
			return
		}
		d, err := server.EnsureDevice([]fn.Option{
			fn.MakeOption("deviceName", devName),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.WriteDeviceContext(d); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(fmt.Sprintf("Account Context %s and Device Context %s created", a.Metadata.Name, d))
	},
}

func init() {
	switchCmd.Flags().StringP("account", "a", "", "account name")
}
