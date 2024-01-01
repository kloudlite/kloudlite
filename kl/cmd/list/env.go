package list

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/server"
	common_util "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var envsCmd = &cobra.Command{
	Use:   "envs",
	Short: "list all the environments accessible to you",
	Long: `List Environments

This command will help you to see list of all the environments that's accessible to you.

Examples:
  # list environments accessible to you
  kl list envs
`,
	Run: func(_ *cobra.Command, _ []string) {
		err := listClusters()
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}

func listEnvironments() error {
	clusters, err := server.GetClusters()

	if err != nil {
		return err
	}

	if len(clusters) == 0 {
		return errors.New("no clusters found")
	}

	clusterName, _ := server.CurrentClusterName()

	header := table.Row{table.HeaderText("name"), table.HeaderText("id"), table.HeaderText("ready")}
	rows := make([]table.Row, 0)

	for _, a := range clusters {
		rows = append(rows, table.Row{
			func() string {
				if a.Metadata.Name == clusterName {
					return text.Colored(fmt.Sprint("*", a.DisplayName), 2)
				}
				return a.DisplayName
			}(),
			func() string {
				if a.Metadata.Name == clusterName {
					return text.Colored(a.Metadata.Name, 2)
				}
				return a.Metadata.Name
			}(),

			func() string {
				if a.Metadata.Name == clusterName {
					return text.Colored(a.Status.IsReady, 2)
				}
				return fmt.Sprint(a.Status.IsReady)
			}(),
		})
	}

	fmt.Println(table.Table(&header, rows))
	table.TotalResults(len(clusters), true)

	return nil
}

func init() {
	clustersCmd.Aliases = append(clustersCmd.Aliases, "enviroments")
	clustersCmd.Aliases = append(clustersCmd.Aliases, "env")
	clustersCmd.Aliases = append(clustersCmd.Aliases, "envs")
}
