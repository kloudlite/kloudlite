/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package use

import (
	"errors"

	"github.com/kloudlite/kl/lib"
	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var accountsCmd = &cobra.Command{
	Use:   "account",
	Short: "select account to use later with all commands",
	Long: `Select account
Examples:
  # select account
  kl use account

	# this will open selector where you can select one of the account accessible to you.

  # select account with account id
  kl use account <accountId>
	`,
	Run: func(_ *cobra.Command, args []string) {
		accountId, err := SelectAccount(args)

		if err != nil {
			common.PrintError(err)
			return
		}

		err = lib.SelectAccount(accountId)
		if err != nil {
			common.PrintError(err)
			return
		}

	},
}

func SelectAccount(args []string) (string, error) {
	accountId := ""
	if len(args) >= 1 {
		accountId = args[0]
	}
	accounts, err := server.GetAccounts()
	if err != nil {
		return "", err
	}

	if accountId != "" {
		for _, a := range accounts {
			if a.Id == accountId {
				return a.Id, nil
			}
		}
		return "", errors.New("you don't have access to this account")
	}

	selectedIndex, err := fuzzyfinder.Find(
		accounts,
		func(i int) string {
			return accounts[i].Name
		},
		fuzzyfinder.WithPromptString("Use Account >"),
	)

	if err != nil {
		return "", err
	}

	return accounts[selectedIndex].Id, nil
}
