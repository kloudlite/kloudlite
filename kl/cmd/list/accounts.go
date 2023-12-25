package list

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/table"
	"github.com/kloudlite/kl/lib/common/ui/text"
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
		err := listAccounts()
		if err != nil {
			common.PrintError(err)
			return
		}
	},
}

func listAccounts() error {
	accounts, err := server.GetAccounts()

	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		return errors.New("no accounts found")
	}

	accountName, _ := server.CurrentAccountName()

	header := table.Row{table.HeaderText("name"), table.HeaderText("id")}
	rows := make([]table.Row, 0)

	for _, a := range accounts {
		rows = append(rows, table.Row{
			func() string {
				if a.Metadata.Name == accountName {
					return text.Colored(fmt.Sprint("*", a.DisplayName), 2)
				}
				return a.DisplayName
			}(),

			func() string {
				if a.Metadata.Name == accountName {
					return text.Colored(a.Metadata.Name, 2)
				}
				return a.Metadata.Name
			}(),
		})
	}

	fmt.Println(table.Table(&header, rows))
	table.TotalResults(len(accounts), true)

	return nil
}

func init() {
	accountsCmd.Aliases = append(accountsCmd.Aliases, "account")
	accountsCmd.Aliases = append(accountsCmd.Aliases, "acc")
}
