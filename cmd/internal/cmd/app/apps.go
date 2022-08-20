package app

import (
	"fmt"
	"log"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/lib/server"
)

// TODO:depricated for now
// appsCmd represents the apps command
var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
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
