package list

import (
	"fmt"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"

	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var configsCmd = &cobra.Command{
	Use:   "configs",
	Short: "Get list of configs in selected environment",
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
		currentTeam, err := fc.CurrentTeamName()
		if err != nil {
			fn.PrintError(err)
			return
		}
		currentEnv, err := apic.EnsureEnv()
		if err != nil {
			fn.PrintError(err)
			return
		}
		config, err := apic.ListConfigs(currentTeam, currentEnv.Name)

		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printConfigs(apic, cmd, config); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printConfigs(apic apiclient.ApiClient, cmd *cobra.Command, configs []apiclient.Config) error {
	e, err := apic.EnsureEnv()
	if err != nil {
		return fn.NewE(err)
	}

	if len(configs) == 0 {
		return fn.Errorf("[#] no configs found in environemnt: %s", text.Blue(e.Name))
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

	fn.Println(table.Table(&header, rows, cmd))

	table.TotalResults(len(configs), true)
	return nil
}

func init() {
	configsCmd.Aliases = append(configsCmd.Aliases, "config")
	configsCmd.Aliases = append(configsCmd.Aliases, "conf")

	configsCmd.Flags().StringP("env", "e", "", "environment name")
}
