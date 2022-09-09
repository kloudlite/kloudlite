package list

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/color"
	"github.com/kloudlite/kl/lib/common/ui/table"
	"github.com/kloudlite/kl/lib/server"
	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "list all the projects accessible to you in selected account",
	Long: `list projects
Examples:
  # list all the projects present in selected account
  kl list projects

	# list all the projects in specific account 
	kl list projects <accountId>

Note: selected project will be highlighted with green color.
  `,
	Run: func(_ *cobra.Command, args []string) {
		accountId := ""
		if len(args) >= 1 {
			accountId = args[0]
		}

		err := listProjects(accountId)
		if err != nil {
			common.PrintError(err)
			return
		}
	},
}

func listProjects(accountId string) error {
	var projects []server.Project
	var err error

	if accountId != "" {
		projects, err = server.GetProjects(common.MakeOption("accountId", accountId))
	} else {
		projects, err = server.GetProjects()
	}

	if err != nil {
		return err
	}

	if len(projects) == 0 {
		return errors.New("no projects found")
	}

	header := table.Row{
		table.HeaderText("projects"),
		table.HeaderText("id"),
	}

	rows := make([]table.Row, 0)

	projectId, _ := server.CurrentProjectId()

	for _, a := range projects {
		rows = append(rows, table.Row{
			func() string {
				if a.Id == projectId {
					return color.ColorText(fmt.Sprint("*", a.Name), 2)
				}
				return a.Name
			}(),

			func() string {
				if a.Id == projectId {
					return color.ColorText(a.Id, 2)
				}
				return a.Id
			}(),
		})
	}

	fmt.Println(table.Table(header, rows))

	if accountId == "" {
		accountId, _ = server.CurrentAccountId()
	}

	if accountId == "" {
		fmt.Println(table.KVOutput("projects of", accountId))
	}

	fmt.Println(table.TotalResults(len(projects)))

	return nil
}

func init() {
	projectsCmd.Aliases = append(projectsCmd.Aliases, "project")
	projectsCmd.Aliases = append(projectsCmd.Aliases, "proj")
}
