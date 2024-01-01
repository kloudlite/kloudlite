package list

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/kloudlite/kl/lib/ui/table"
	"github.com/kloudlite/kl/lib/ui/text"
	"github.com/kloudlite/kl/lib/util"
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
		accountName := ""
		if len(args) >= 1 {
			accountName = args[0]
		}

		err := listProjects(accountName)
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}

func listProjects(accountName string) error {
	var projects []server.Project
	var err error

	if accountName != "" {
		projects, err = server.ListProjects(common_util.MakeOption("accountName", accountName))
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
		table.HeaderText("projects"),
		table.HeaderText("id"),
	}

	rows := make([]table.Row, 0)

	projectId, _ := server.CurrentProjectId()

	for _, a := range projects {
		rows = append(rows, table.Row{
			func() string {
				if a.Id == projectId {
					return text.Colored(fmt.Sprint("*", a.Name), 2)
				}
				return a.Name
			}(),

			func() string {
				if a.Id == projectId {
					return text.Colored(a.Id, 2)
				}
				return a.Id
			}(),
		})
	}

	fmt.Println(table.Table(&header, rows))

	if accountName == "" {
		accountName, _ = util.CurrentAccountName()
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
