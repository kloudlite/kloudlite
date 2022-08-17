/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package account

import (
	"fmt"
	"github.com/ktr0731/go-fuzzyfinder"
	"kloudlite.io/cmd/internal/lib"
	"kloudlite.io/cmd/internal/lib/server"
	"log"

	"github.com/spf13/cobra"
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
			TriggerSelectAccount()
			return
		}
		lib.SelectAccount(args[0])
	},
}

func TriggerSelectAccount() {
	accounts, err := server.GetAccounts()
	if err != nil {
		log.Fatal(err)
	}
	selectedIndex, err := fuzzyfinder.Find(
		accounts,
		func(i int) string {
			return accounts[i].Name
		},
		fuzzyfinder.WithPromptString("Use Account >"),
	)

	if err != nil {
		log.Fatal(err)
	}
	lib.SelectAccount(accounts[selectedIndex].Id)
	fmt.Println("Using account: " + accounts[selectedIndex].Name)
}

func init() {
	Cmd.AddCommand(useCmd)
}
