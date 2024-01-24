package sw

import (
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var accCmd = &cobra.Command{
	Use:   "account",
	Short: "Switch account",
	Long: `Use this command to switch account
Example:
  # switch to a different account
  kl switch account
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		accountName := fn.ParseStringFlag(cmd, "account")

		acc, err := server.SelectAccount(accountName)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.WriteAccountContext(acc.Metadata.Name); err != nil {
			fn.PrintError(err)
			return
		}

		d, err := server.EnsureDevice([]fn.Option{
			fn.MakeOption("accountName", acc.Metadata.Name),
		}...)

		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.WriteDeviceContext(d); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	accCmd.Flags().StringP("account", "a", "", "account name")
	accCmd.Aliases = append(accCmd.Aliases, "acc")
}
