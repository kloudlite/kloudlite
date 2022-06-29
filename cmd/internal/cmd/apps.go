/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/lib/server"
	"log"
	"time"
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
	s := spinner.New(spinner.CharSets[31], 100*time.Millisecond)
	s.Start()
	apps, err := server.GetApps()
	s.Stop()
	if err != nil {
		log.Fatal(err)
	}
	selectedIndex, err := fuzzyfinder.Find(
		apps,
		func(i int) string {
			return apps[i].Name
		},
		fuzzyfinder.WithPromptString("Select Project >"),
	)
	return apps[selectedIndex].Id
}

func init() {
	rootCmd.AddCommand(appsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// appsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// appsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
