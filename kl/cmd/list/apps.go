package list

import (
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Get list of apps in selected environment",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listapps(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listapps(cmd *cobra.Command, _ []string) error {

	envName := fn.ParseStringFlag(cmd, "env")

	filePath := fn.ParseKlFile(cmd)
	klFile, err := client.GetKlFile(filePath)
	if err != nil {
		return functions.NewE(err)
	}

	apps, err := server.ListApps([]fn.Option{
		fn.MakeOption("accountName", klFile.AccountName),
		fn.MakeOption("envName", envName),
	}...)
	if err != nil {
		return functions.NewE(err)
	}

	if len(apps) == 0 {
		return functions.Error("no apps found")
	}

	header := table.Row{
		table.HeaderText("apps"),
		table.HeaderText("name"),
	}

	rows := make([]table.Row, 0)

	for _, a := range apps {
		rows = append(rows, table.Row{a.DisplayName, a.Metadata.Name})
	}

	fn.Println(table.Table(&header, rows, cmd))

	table.KVOutput("apps of", envName, true)
	table.TotalResults(len(apps), true)
	return nil
}

func init() {
	appsCmd.Aliases = append(appsCmd.Aliases, "app")
	appsCmd.Flags().StringP("env", "e", "", "environment name")
}
