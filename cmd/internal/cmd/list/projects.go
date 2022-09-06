package list

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/common/ui/table"
	"kloudlite.io/cmd/internal/lib/server"
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

	colorReset := "\033[0m"
	colorGreen := "\033[32m"

	for _, a := range projects {
		rows = append(rows, table.Row{fmt.Sprintf("%s%s%s", func() string {
			if a.Id == projectId {
				return colorGreen
			}
			return ""
		}(), a.Name, func() string {
			if a.Id == projectId {
				return colorReset
			}
			return ""
		}()), fmt.Sprintf("%s%s%s",
			func() string {
				if a.Id == projectId {
					return colorGreen
				}
				return ""
			}(),
			a.Id,
			func() string {
				if a.Id == projectId {
					return colorReset
				}
				return ""
			}(),
		)})

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
}
