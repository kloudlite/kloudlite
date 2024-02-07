package list

import (
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var configsCmd = &cobra.Command{
	Use:   "configs",
	Short: "Get list of configs in current project & selected environment",
	Run: func(cmd *cobra.Command, args []string) {

		pName := fn.ParseStringFlag(cmd, "project")
		envName := fn.ParseStringFlag(cmd, "env")

		config, err := server.ListConfigs([]fn.Option{
			fn.MakeOption("projectName", pName),
			fn.MakeOption("envName", envName),
		}...)

		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printConfigs(cmd, config); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printConfigs(cmd *cobra.Command, configs []server.Config) error {

	e, err := client.CurrentEnv()
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		return fmt.Errorf("[#] no configs found in environemnt: %s", text.Blue(e.Name))
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

	fmt.Println(table.Table(&header, rows, cmd))

	if s := fn.ParseStringFlag(cmd, "output"); s == "table" {
		pName, _ := client.CurrentProjectName()
		if pName != "" {
			table.KVOutput("configs of", pName, true)
		}
	}
	table.TotalResults(len(configs), true)
	return nil
}

func init() {
	configsCmd.Aliases = append(configsCmd.Aliases, "config")
	configsCmd.Aliases = append(configsCmd.Aliases, "conf")

	configsCmd.Flags().StringP("project", "p", "", "project name")
	configsCmd.Flags().StringP("env", "e", "", "environment name")

	fn.WithOutputVariant(configsCmd)
}
