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

var lambdasCmd = &cobra.Command{
	Use:   "lambdas",
	Short: "list all the lambdas available in selected project",
	Long: `List all the lambdas available in project.
Examples:
	# list all the lambdas with selected project
  kl list lambdas

	# list all the lambdas with projectId
  kl list lambdas <projectId>
`,
	Run: func(_ *cobra.Command, args []string) {
		err := listlambdas(args)
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}

func listlambdas(args []string) error {
	var l []server.App
	var err error

	projectId := ""

	if len(args) >= 1 {
		projectId = args[0]
	}

	if projectId == "" {
		l, err = server.GetApps()
	} else {
		l, err = server.GetApps(common_util.MakeOption("projectId", args[0]))
	}

	var lambdas []server.App
	for _, a := range l {
		if a.IsLambda {
			lambdas = append(lambdas, a)
		}
	}

	if err != nil {
		return err
	}

	if len(lambdas) == 0 {
		return errors.New("no lambdas found")
	}

	header := table.Row{table.HeaderText("apps"), table.HeaderText("id")}
	rows := make([]table.Row, 0)

	for _, a := range lambdas {
		rows = append(rows, table.Row{a.Name})
	}

	fmt.Println(table.Table(&header, rows))

	if projectId == "" {
		projectId, err = client.CurrentProjectName()
		if err != nil {
			return err
		}
	}

	table.KVOutput("apps of", projectId, true)
	table.TotalResults(len(lambdas), true)

	return nil
}

func init() {
	lambdasCmd.Aliases = append(lambdasCmd.Aliases, "lambda")
}
