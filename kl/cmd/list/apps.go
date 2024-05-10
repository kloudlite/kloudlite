package list

import (
	"errors"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Get list of apps in current project & selected environment",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listapps(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listapps(cmd *cobra.Command, _ []string) error {

	envName := fn.ParseStringFlag(cmd, "env")
	accName := fn.ParseStringFlag(cmd, "account")

	apps, err := server.ListApps([]fn.Option{
		fn.MakeOption("accountName", accName),
		fn.MakeOption("envName", envName),
	}...)
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		return errors.New("no apps found")
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
	fn.WithOutputVariant(appsCmd)
}
