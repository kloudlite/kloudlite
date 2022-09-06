package use

import (
	"errors"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/lib"
	"kloudlite.io/cmd/internal/lib/server"
)

// listCmd represents the list command
var projectsCmd = &cobra.Command{
	Use:   "project",
	Short: "select project to use later with all commands",
	Long: `Select project
Examples:
  # select project
  kl use project

	# this will open selector where you can select one of the project accessible to you in selected account context.

  # select project with project id
  kl use project <projectId>

  # select project with project readableId
  kl use project <project_readableId>
`,
	Run: func(_ *cobra.Command, args []string) {
		projectId, err := SelectProject(args)
		if err != nil {
			common.PrintError(err)
			return
		}

		err = lib.SelectProject(projectId)
		if err != nil {
			common.PrintError(err)
			return
		}

	},
}

func SelectProject(args []string) (string, error) {
	projectId := ""

	if len(args) >= 1 {
		projectId = args[0]
	}

	projects, err := server.GetProjects()
	if err != nil {
		return "", err
	}

	if projectId != "" {

		for _, p := range projects {
			if p.Id == projectId || p.ReadableId == projectId {
				return p.Id, nil
			}
		}

		return "", errors.New("no projects found with the provided id in selected account")
	}

	selectedIndex, err := fuzzyfinder.Find(
		projects,
		func(i int) string {
			return projects[i].Name
		},
		fuzzyfinder.WithPromptString("Select Project >"),
	)
	if err != nil {
		return "", err
	}

	return projects[selectedIndex].Id, nil
}
