package list

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var configsCmd = &cobra.Command{
	Use:   "configs",
	Short: "list all the configs available in selected project",
	Long: `List all the configs available in project.

Examples:

This command will provide the list of all the configs for the selected project.
  kl list configs

This command will provide the list of all the configs for the provided project name.
  kl list configs <projectName>
`,
	Run: func(_ *cobra.Command, args []string) {
		pName := ""
		if len(args) > 1 {
			pName = args[0]
		}

		config, err := server.ListConfigs(fn.MakeOption("projectName", pName))
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printConfigs(config); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printConfigs(configs []server.Config) error {
	if len(configs) == 0 {
		return errors.New("no configs found")
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		table.HeaderText("entries"),
	}

	rows := make([]table.Row, 0)

	for _, a := range configs {
		rows = append(rows, table.Row{a.DisplayName, a.Metadata.Name, fmt.Sprintf("%d", len(a.Data))})
	}

	fmt.Println(table.Table(&header, rows))

	pName, _ := client.CurrentProjectName()

	if pName != "" {
		table.KVOutput("configs of", pName, true)
	}

	table.TotalResults(len(configs), true)

	return nil
}

func init() {
	configsCmd.Aliases = append(configsCmd.Aliases, "config")
	configsCmd.Aliases = append(configsCmd.Aliases, "conf")
}
