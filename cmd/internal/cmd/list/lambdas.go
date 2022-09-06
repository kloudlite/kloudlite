package list

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/common/ui/table"
	"kloudlite.io/cmd/internal/lib/server"
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
			common.PrintError(err)
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
		l, err = server.GetApps(common.MakeOption("projectId", args[0]))
	}

	lambdas := []server.App{}
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

	fmt.Println(table.Table(header, rows))

	if projectId == "" {
		projectId, err = server.CurrentProjectId()
		if err != nil {
			return err
		}
	}

	fmt.Println(table.KVOutput("apps of", projectId))
	fmt.Println(table.TotalResults(len(lambdas)))

	return nil
}

func init() {
	lambdasCmd.Aliases = append(lambdasCmd.Aliases, "lambda")
}
