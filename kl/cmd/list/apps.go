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

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "list all the apps available in selected project",
	Long: `List all the apps available in project.

This command will provide the list of all the apps that's accessible for the selected environment or workspace.

Examples:
  kl list apps

	`,
	Run: func(_ *cobra.Command, args []string) {
		if err := listapps(args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listapps(args []string) error {
	var apps []server.App
	var err error

	projectId := ""

	if len(args) >= 1 {
		projectId = args[0]
	}

	if projectId == "" {
		apps, err = server.ListApps()
	} else {
		apps, err = server.ListApps(fn.MakeOption("projectId", args[0]))
	}

	if err != nil {
		return err
	}

	if len(apps) == 0 {
		return errors.New("no apps found")
	}

	header := table.Row{
		table.HeaderText("apps"),
		table.HeaderText("id"),
	}

	rows := make([]table.Row, 0)

	for _, a := range apps {
		rows = append(rows, table.Row{a.DisplayName, a.Metadata.Name})
	}

	fmt.Println(table.Table(&header, rows))

	if projectId == "" {
		projectId, err = client.CurrentProjectName()
		if err != nil {
			return err
		}
	}

	table.KVOutput("apps of", projectId, true)
	table.TotalResults(len(apps), true)

	return nil
}

func init() {
	appsCmd.Aliases = append(appsCmd.Aliases, "app")
}
