package list

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/common/ui/table"
	"kloudlite.io/cmd/internal/lib/server"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "list all the apps available in selected project",
	Long: `List all the apps available in project.
Examples:
	# list all the apps with selected project
  kl list apps

	# list all the apps with projectId
  kl list apps <projectId>
	`,
	Run: func(_ *cobra.Command, args []string) {
		err := listapps(args)
		if err != nil {
			common.PrintError(err)
			return
		}
	},
}

func listapps(args []string) error {
	var a []server.App
	var err error

	projectId := ""

	if len(args) >= 1 {
		projectId = args[0]
	}

	if projectId == "" {
		a, err = server.GetApps()
	} else {
		a, err = server.GetApps(common.MakeOption("projectId", args[0]))
	}

	apps := []server.App{}
	for _, l := range a {
		if !l.IsLambda {
			apps = append(apps, l)
		}
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
		rows = append(rows, table.Row{a.Name, a.Id})
	}

	fmt.Println(table.Table(header, rows))

	if projectId == "" {
		projectId, err = server.CurrentProjectId()
		if err != nil {
			return err
		}
	}

	fmt.Println(table.KVOutput("apps of", projectId))
	fmt.Println(table.TotalResults(len(apps)))

	return nil
}

func init() {
	appsCmd.Aliases = append(appsCmd.Aliases, "app")
}
