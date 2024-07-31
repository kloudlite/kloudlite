package list

import (
	"fmt"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/kloudlite/kl/domain/fileclient"

	"github.com/kloudlite/kl/domain/apiclient"
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

		currentEnv, err := apic.EnsureEnv()
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

		if err := printMres(apic, cmd, mres); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printMres(apic apiclient.ApiClient, cmd *cobra.Command, mres []apiclient.Mres) error {
	e, err := apic.EnsureEnv()
	if err != nil {
		return fn.NewE(err)
	}
	if len(mres) == 0 {
		return fmt.Errorf("[#] no managed resources found in environemnt: %s", text.Blue(e.Name))
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		table.HeaderText("Secret Ref Name"),
	}

	rows := make([]table.Row, 0)

	for _, a := range mres {
		rows = append(rows, table.Row{a.DisplayName, a.Name, a.SecretRefName.Name})
	}

	fmt.Println(table.Table(&header, rows))
	table.TotalResults(len(mres), true)
	return nil
}

func init() {
	mresCmd.Aliases = append(mresCmd.Aliases, "mres")
	mresCmd.Aliases = append(mresCmd.Aliases, "managed-resources")
	mresCmd.Aliases = append(mresCmd.Aliases, "mresources")
	fn.WithOutputVariant(mresCmd)
}
