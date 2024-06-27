package list

import (
	"fmt"

	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var mresCmd = &cobra.Command{
	Use:   "mreses",
	Short: "Get list of managed resources in selected environment",
	Run: func(cmd *cobra.Command, args []string) {

		sec, err := server.ListMreses()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printMres(cmd, sec); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printMres(_ *cobra.Command, secrets []server.Mres) error {
	if len(secrets) == 0 {
		return functions.Error("no secrets found")
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		// table.HeaderText("entries"),
	}

	rows := make([]table.Row, 0)

	for _, a := range secrets {
		rows = append(rows, table.Row{a.DisplayName, a.Metadata.Name})
	}

	fmt.Println(table.Table(&header, rows))
	table.TotalResults(len(secrets), true)
	return nil
}

func init() {
	mresCmd.Aliases = append(secretsCmd.Aliases, "mres")
	mresCmd.Aliases = append(secretsCmd.Aliases, "managed-resources")
	mresCmd.Aliases = append(secretsCmd.Aliases, "mresources")
	fn.WithOutputVariant(mresCmd)
}
