package list

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/color"
	"github.com/kloudlite/kl/lib/common/ui/table"
	"github.com/kloudlite/kl/lib/server"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "list all the accounts accessible to you",
	Long: `List Accounts

This command will help you to see list of all the accounts that's accessible to you. 

Examples:
  # list accounts accessible to you
  kl list accounts
`,
	Run: func(_ *cobra.Command, _ []string) {
		err := listAcocunts()
		if err != nil {
			common.PrintError(err)
			return
		}
	},
}

func listAcocunts() error {
	accounts, err := server.GetAccounts()

	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		return errors.New("no accounts found")
	}

	accountId, _ := server.CurrentAccountId()

	header := table.Row{table.HeaderText("accounts"), table.HeaderText("id")}
	rows := make([]table.Row, 0)

	for _, a := range accounts {
		rows = append(rows, table.Row{
			func() string {
				if a.Id == accountId {
					return color.ColorText(fmt.Sprint("*", a.Name), 2)
				}
				return a.Name
			}(),

			func() string {
				if a.Id == accountId {
					return color.ColorText(a.Id, 2)
				}
				return a.Id
			}(),
		})
	}

	fmt.Println(table.Table(header, rows))
	fmt.Println(table.TotalResults(len(accounts)))

	return nil
}

func init() {
	accountsCmd.Aliases = append(accountsCmd.Aliases, "account")
	accountsCmd.Aliases = append(accountsCmd.Aliases, "acc")
}
