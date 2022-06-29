/*
Copyright © 2022 Kloudlite <support@kloudlite.io>

*/
package cmd

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/lib"
	"kloudlite.io/cmd/internal/lib/server"
	"log"
	"time"
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
	Run: func(cmd *cobra.Command, args []string) {
		TriggerSelectAccount()
	},
}

func TriggerSelectAccount() {
	s := spinner.New(spinner.CharSets[31], 100*time.Millisecond)

	s.Start()
	accounts, err := server.GetAccounts()
	s.Stop()
	if err != nil {
		log.Fatal(err)
	}
	selectedIndex, err := fuzzyfinder.Find(
		accounts,
		func(i int) string {
			return accounts[i].Name
		},
		fuzzyfinder.WithPromptString("Select Account >"),
	)

	if err != nil {
		log.Fatal(err)
	}
	lib.SelectAccount(accounts[selectedIndex].Id)
	fmt.Println("Selected account: " + accounts[selectedIndex].Name)
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
