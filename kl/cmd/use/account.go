package use

import (
	"fmt"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

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

  # select account with account name
  kl use account <accountname>
	`,
	Run: func(_ *cobra.Command, args []string) {

		aName := ""

		if len(args) >= 1 {
			aName = args[0]
		}

		account, err := server.SelectAccount(aName)

		if err != nil {
			fn.PrintError(err)
			return
		}

		fmt.Println(text.Bold(text.Green("\nSelected account:")),
			text.Blue(fmt.Sprintf("%s (%s)", account.DisplayName, account.Metadata.Name)),
		)
	},
}
