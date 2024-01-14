package context

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "listing all infra contexts",
	Long: `This command let you list all infra contexts.
Example:
  # list all infra contexts
  kl infra context list
	`,
	Run: func(_ *cobra.Command, _ []string) {

		c, err := client.GetInfraContexts()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if len(c.InfraContexts) == 0 {
			fn.PrintError(fmt.Errorf("no infra context found"))
			return
		}

		header := table.Row{
			table.HeaderText("Name"),
			table.HeaderText("Account_Name"),
			table.HeaderText("Cluster_Name"),
			table.HeaderText("Device_Name"),
		}

		rows := make([]table.Row, 0)
		for _, ctx := range c.InfraContexts {
			rows = append(rows, table.Row{
				fn.GetPrintRow2(ctx.Name, ctx.Name == c.ActiveContext, true),
				fn.GetPrintRow2(ctx.AccountName, ctx.Name == c.ActiveContext),
				fn.GetPrintRow2(ctx.ClusterName, ctx.Name == c.ActiveContext),
				fn.GetPrintRow2(ctx.DeviceName, ctx.Name == c.ActiveContext),
			})
		}

		fmt.Println(table.Table(&header, rows))

		table.TotalResults(len(c.InfraContexts), true)
	},
}

func init() {
	listCmd.Aliases = append(listCmd.Aliases, "ls")
}
