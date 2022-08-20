/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package account

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all the accounts accessible to you and select one.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		TriggerSelectAccount()
	},
}

func init() {
	Cmd.AddCommand(listCmd)
}
