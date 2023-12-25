package use

import (
	"fmt"

	"github.com/kloudlite/kl/cmd/util"
	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/text"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var accountsCmd = &cobra.Command{
	Use:   "account",
	Short: "select account to use later with all commands",
	Long: `Select account
Examples:
  # select account
  kl use account
	# this will open selector where you can select one of the account accessible to you.

  # select account with account id
  kl use account <accountId>
	`,
	Run: func(_ *cobra.Command, args []string) {
		accountName, err := util.SelectAccount(args)

		if err != nil {
			common.PrintError(err)
			return
		}

		fmt.Println(text.Bold(text.Green("\nSelected account:")), text.Blue(accountName))
	},
}
