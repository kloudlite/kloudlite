/*
Copyright © 2022 Kloudlite <support@kloudlite.io>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

// accountsCmd represents the accounts command
var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	TriggerSelectAccount()
	// },
}

func init() {
	rootCmd.AddCommand(accountsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// accountsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
