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

var envsCmd = &cobra.Command{
	Use:   "envs",
	Short: "list all the environments accessible to you",
	Long: `List Environments

This command will provide the list of all the environments that's accessible to you. 

Examples:
  # list environments accessible to you
  kl list envs

Note: selected project will be highlighted with green color.

`,
	Run: func(_ *cobra.Command, args []string) {
		err := listEnvironments(args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listEnvironments(args []string) error {

	pName := ""
	if len(args) >= 1 {
		pName = args[0]
	}

	envs, err := server.ListEnvs(fn.MakeOption("projectName", pName))
	if err != nil {
		return err
	}

	if len(envs) == 0 {
		return errors.New("no environments found")
	}

	env, _ := client.CurrentEnv()
	envName := ""
	if env != nil {
		envName = env.Name
	}

	header := table.Row{table.HeaderText("DisplayName"), table.HeaderText("Name"), table.HeaderText("ready")}
	rows := make([]table.Row, 0)

	for _, a := range envs {
		rows = append(rows, table.Row{
			fn.GetPrintRow(a, envName, a.DisplayName, true),
			fn.GetPrintRow(a, envName, a.Metadata.Name),
			fn.GetPrintRow(a, envName, a.Status.IsReady),
		})
	}

	fmt.Println(table.Table(&header, rows))
	table.TotalResults(len(envs), true)

	return nil
}

func init() {
	envsCmd.Aliases = append(clustersCmd.Aliases, "enviroments")
	envsCmd.Aliases = append(clustersCmd.Aliases, "env")
	envsCmd.Aliases = append(clustersCmd.Aliases, "envs")
}
