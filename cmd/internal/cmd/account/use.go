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
	Short: "use account by providing account_id directly",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			TriggerSelectAccount()
			return
		}
		accounts, err := server.GetAccounts()
		if err != nil {
			fmt.Println(err)
			return
		}
		found := false

		for _, a := range accounts {
			if args[0] == a.Id {
				found = true
			}
		}

		if found {
			lib.SelectAccount(args[0])
		} else {
			fmt.Println("You don't have access to this account")
		}

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
