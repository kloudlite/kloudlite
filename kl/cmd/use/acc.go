package use

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var accCmd = &cobra.Command{
	Use:   "account",
	Short: "Switch account",
	Run: func(cmd *cobra.Command, args []string) {
		accountName := fn.ParseStringFlag(cmd, "account")

		acc, err := server.SelectAccount(accountName)
		if err != nil {
			fn.PrintError(err)
			return
		}

		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.StopAll(); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Logf("%s %s\n", text.Blue(text.Bold("\nSelected Account:")), acc.Metadata.Name)
	},
}

func init() {
	accCmd.Flags().StringP("account", "a", "", "account name")
	accCmd.Aliases = append(accCmd.Aliases, "acc")
}
