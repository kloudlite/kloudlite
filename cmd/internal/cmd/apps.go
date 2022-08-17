package cmd

import (
	"fmt"
	"log"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/lib/server"
)

// appsCmd represents the apps command
var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		appId := TriggerSelectApp()
		fmt.Println("Selected App: " + appId)
	},
}

func TriggerSelectApp() string {

	apps, err := server.GetApps()
	if err != nil {
		log.Fatal(err)
	}
	selectedIndex, err := fuzzyfinder.Find(
		apps,
		func(i int) string {
			return apps[i].Name
		},
		fuzzyfinder.WithPromptString("Select App >"),
	)
	if err != nil {
		log.Fatal(err)
	}

	return apps[selectedIndex].Id
}

func init() {
	rootCmd.AddCommand(appsCmd)
}
