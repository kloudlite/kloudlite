package list

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	common_util "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var regionsCmd = &cobra.Command{
	Use:   "regions",
	Short: "list all regions of selected device",
	Long: `List regions

This command will help you to see list of all the regions of your selected device. 

Examples:
  # list regions of selected device
  kl list regions
`,
	Run: func(_ *cobra.Command, _ []string) {
		err := listRegions()
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}

func listRegions() error {
	regions, err := server.GetRegions()
	if err != nil {
		return err
	}

	if len(regions) == 0 {
		return errors.New("no regions available in current account")
	}

	header := table.Row{
		table.HeaderText("Provider"),
		table.HeaderText("name"),
		table.HeaderText("region"),
	}

	rows := make([]table.Row, 0)

	for _, r := range regions {
		rows = append(rows, table.Row{
			r.Provider, r.Name, r.Region,
		})
	}

	fmt.Println(table.Table(&header, rows))

	if accountId, _ := client.CurrentAccountName(); accountId != "" {
		table.KVOutput("regions of", accountId, true)
	}

	return nil
}

func init() {
	regionsCmd.Aliases = append(accountsCmd.Aliases, "region")
	regionsCmd.Aliases = append(accountsCmd.Aliases, "reg")
}
