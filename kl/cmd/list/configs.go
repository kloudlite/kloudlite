package list

import (
	"fmt"

	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var configsCmd = &cobra.Command{
	Use:   "configs",
	Short: "Get list of configs in selected environment",
	Run: func(cmd *cobra.Command, args []string) {

		envName := fn.ParseStringFlag(cmd, "env")

		filePath := fn.ParseKlFile(cmd)
		klFile, err := fileclient.GetKlFile(filePath)
		if err != nil {
			fn.PrintError(err)
			return
		}
		config, err := apiclient.ListConfigs([]fn.Option{
			fn.MakeOption("envName", envName),
			fn.MakeOption("accountName", klFile.AccountName),
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

func printConfigs(cmd *cobra.Command, configs []apiclient.Config) error {

	e, err := fileclient.CurrentEnv()
	if err != nil {
		return functions.NewE(err)
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

	table.TotalResults(len(configs), true)
	return nil
}

func init() {
	configsCmd.Aliases = append(configsCmd.Aliases, "config")
	configsCmd.Aliases = append(configsCmd.Aliases, "conf")

	configsCmd.Flags().StringP("env", "e", "", "environment name")
}
