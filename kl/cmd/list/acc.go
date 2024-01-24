package list

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var accCmd = &cobra.Command{
	Use:   "accounts",
	Short: "list all the accounts accessible to you",
	Long: `List Accounts

This command will provide the list of all the accounts that's accessible to you. 

Examples:
  # list accounts accessible to you
  kl account list

Note: selected account will be highlighted with green color.

`,
	Run: func(cmd *cobra.Command, _ []string) {
		err := listAccounts(cmd)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listAccounts(cmd *cobra.Command) error {
	accounts, err := server.ListAccounts()

	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		return errors.New("no accounts found")
	}

	accountName, _ := client.CurrentAccountName()

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

	fmt.Println(table.Table(&header, rows, cmd))

	table.TotalResults(len(rows), true)
	return nil
}

func init() {
	accCmd.Aliases = append(accCmd.Aliases, "acc", "account")

	fn.WithOutputVariant(accCmd)
}
