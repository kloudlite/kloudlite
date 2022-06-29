/*
Copyright © 2022 Kloudlite <support@kloudlite.io>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/lib/server"
)

// loadenvCmd represents the loadenv command
var loadenvCmd = &cobra.Command{
	Use:   "loadenv",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		appId := TriggerSelectApp()
		app, err := server.GetApp(appId)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(app)
	},
}

func init() {
	rootCmd.AddCommand(loadenvCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadenvCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadenvCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
