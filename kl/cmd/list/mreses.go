package list

import (
	"fmt"

	"github.com/kloudlite/kl/domain/fileclient"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var mresCmd = &cobra.Command{
	Use:   "mreses",
	Short: "Get list of managed resources in selected environment",
	Run: func(cmd *cobra.Command, args []string) {
		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		currentEnv, err := fc.CurrentEnv()
		if err != nil {
			fn.PrintError(err)
			return
		}

		currentAccount, err := fc.CurrentAccountName()
		if err != nil {
			fn.PrintError(err)
			return
		}

		mres, err := apic.ListMreses(currentAccount, currentEnv.Name)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printMres(cmd, mres); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printMres(_ *cobra.Command, secrets []apiclient.Mres) error {
	if len(secrets) == 0 {
		return functions.Error("no managed resources found")
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		table.HeaderText("Secret Ref Name"),
	}

	rows := make([]table.Row, 0)

	for _, a := range secrets {
		rows = append(rows, table.Row{a.DisplayName, a.Name, a.SecretRefName.Name})
	}

	fmt.Println(table.Table(&header, rows))
	table.TotalResults(len(secrets), true)
	return nil
}

func init() {
	mresCmd.Aliases = append(mresCmd.Aliases, "mres")
	mresCmd.Aliases = append(mresCmd.Aliases, "managed-resources")
	mresCmd.Aliases = append(mresCmd.Aliases, "mresources")
	fn.WithOutputVariant(mresCmd)
}
