/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package project

import (
	"fmt"
	"os"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/lib"
	"kloudlite.io/cmd/internal/lib/server"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			TriggerSelectProject()
			return
		}

		projects, err := server.GetProjects()

		if err != nil {
			fmt.Println(err)
			return
		}
		found := false

		for _, p := range projects {
			if args[0] == p.Id {
				found = true
			}
		}

		if found {
			lib.SelectProject(args[0])
		} else {
			fmt.Fprint(os.Stderr, "You don't have access to this project or either it's not present under the selected account")
		}

	},
}

func TriggerSelectProject() {
	projects, err := server.GetProjects()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
	selectedIndex, err := fuzzyfinder.Find(
		projects,
		func(i int) string {
			return projects[i].Name
		},
		fuzzyfinder.WithPromptString("Select Project >"),
	)
	if err != nil {
		common.PrintError(err)
		return
	}
	lib.SelectProject(projects[selectedIndex].Id)
	fmt.Println("Selected project: " + projects[selectedIndex].Name)
}
