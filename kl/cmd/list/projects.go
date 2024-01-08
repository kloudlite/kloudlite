package list

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "list all the projects accessible to you in selected account",
	Long: `list projects
Examples:
This command will provide the list of all the projects in selected account. 
  kl list projects

Note: selected project will be highlighted with green color.
  `,
	Run: func(_ *cobra.Command, args []string) {
		accountName := ""
		if len(args) >= 1 {
			accountName = args[0]
		}

		err := listProjects(accountName)
		if err != nil {
			functions.PrintError(err)
			return
		}
	},
}

func listProjects(accountName string) error {
	var projects []server.Project
	var err error

	if accountName != "" {
		projects, err = server.ListProjects(functions.MakeOption("accountName", accountName))
	} else {
		projects, err = server.ListProjects()
	}

	if err != nil {
		return err
	}

	if len(projects) == 0 {
		return errors.New("no projects found")
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
	}

	rows := make([]table.Row, 0)

	projectId, _ := client.CurrentProjectName()

	for _, a := range projects {
		rows = append(rows, table.Row{
			func() string {
				if a.Metadata.Name == projectId {
					return text.Colored(fmt.Sprint("*", a.DisplayName), 2)
				}
				return a.DisplayName
			}(),

			func() string {
				if a.Metadata.Name == projectId {
					return text.Colored(a.Metadata.Name, 2)
				}
				return a.Metadata.Name
			}(),
		})
	}

	fmt.Println(table.Table(&header, rows))

	if accountName == "" {
		accountName, _ = client.CurrentAccountName()
	}

	if accountName == "" {
		table.KVOutput("projects of", accountName, true)
	}

	table.TotalResults(len(projects), true)

	return nil
}

func init() {
	projectsCmd.Aliases = append(projectsCmd.Aliases, "project")
	projectsCmd.Aliases = append(projectsCmd.Aliases, "proj")
}
