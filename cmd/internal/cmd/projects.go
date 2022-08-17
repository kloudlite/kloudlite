/*
Copyright © 2022 Kloudlite <support@kloudlite.io>

*/
package cmd

import (
	"fmt"
	"log"

	"github.com/ktr0731/go-fuzzyfinder"
	"kloudlite.io/cmd/internal/lib"
	"kloudlite.io/cmd/internal/lib/server"

	"github.com/spf13/cobra"
)

// projectsCmd represents the projects command
var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		TriggerSelectProject()
	},
}

func TriggerSelectProject() {
	projects, err := server.GetProjects()
	if err != nil {
		log.Fatal(err)
	}
	selectedIndex, err := fuzzyfinder.Find(
		projects,
		func(i int) string {
			return projects[i].Name
		},
		fuzzyfinder.WithPromptString("Select Project >"),
	)
	if err != nil {
		log.Fatal(err)
	}
	lib.SelectProject(projects[selectedIndex].Id)
	fmt.Println("Selected project: " + projects[selectedIndex].Name)
}

func init() {
	rootCmd.AddCommand(projectsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// projectsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
