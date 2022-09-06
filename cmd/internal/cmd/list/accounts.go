package list

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/common/ui/table"
	"kloudlite.io/cmd/internal/lib/server"
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

	colorReset := "\033[0m"
	colorGreen := "\033[32m"

	header := table.Row{table.HeaderText("accounts"), table.HeaderText("id")}
	rows := make([]table.Row, 0)

	for _, a := range accounts {
		rows = append(rows, table.Row{fmt.Sprintf("%s%s%s", func() string {
			if a.Id == accountId {
				return colorGreen
			}
			return ""
		}(), a.Name, func() string {
			if a.Id == accountId {
				return colorReset
			}
			return ""

		}()), fmt.Sprintf("%s%s%s",
			func() string {
				if a.Id == accountId {
					return colorGreen
				}
				return ""
			}(),
			a.Id,
			func() string {
				if a.Id == accountId {
					return colorReset
				}
				return ""

			}(),
		)})
	}

	fmt.Println(table.Table(header, rows))
	fmt.Println(table.TotalResults(len(accounts)))

	return nil
}

func init() {
	accountsCmd.Aliases = append(accountsCmd.Aliases, "account")
}
