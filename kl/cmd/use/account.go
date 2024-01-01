package use

import (
	"fmt"
	common_util "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

	common_cmd "github.com/kloudlite/kl/cmd/common"
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
		accountName, err := common_cmd.SelectAccount(args)

		if err != nil {
			common_util.PrintError(err)
			return
		}

		fmt.Println(text.Bold(text.Green("\nSelected account:")),
			text.Blue(fmt.Sprintf("%s (%s)", accountName.DisplayName, accountName.Name)),
		)
	},
}
