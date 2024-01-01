package use

import (
	"errors"
	"github.com/kloudlite/kl/domain/server"
	common_util "github.com/kloudlite/kl/pkg/functions"

	"github.com/kloudlite/kl/lib"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
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
			common_util.PrintError(err)
			return
		}

		err = lib.SelectProject(projectId)
		if err != nil {
			common_util.PrintError(err)
			return
		}

	},
}

func SelectProject(args []string) (string, error) {
	projectName := ""

	if len(args) >= 1 {
		projectName = args[0]
	}

	projects, err := server.ListProjects()
	if err != nil {
		return "", err
	}

	if projectName != "" {

		for _, p := range projects {
			if p.Metadata.Name == projectName {
				return p.DisplayName, nil
			}
		}

		return "", errors.New("no projects found with the provided id in selected account")
	}

	selectedIndex, err := fuzzyfinder.Find(
		projects,
		func(i int) string {
			return projects[i].Metadata.Name
		},
		fuzzyfinder.WithPromptString("Select Project >"),
	)
	if err != nil {
		return "", err
	}

	return projects[selectedIndex].Metadata.Name, nil
}
