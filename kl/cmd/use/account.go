package use

import (
	"github.com/kloudlite/kl/domain/apiclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/k3s"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/spf13/cobra"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "use account",
	Run: func(_ *cobra.Command, _ []string) {
		if err := UseAccount(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func UseAccount() error {
	apic, err := apiclient.New()

	if err != nil {
		return fn.NewE(err)
	}
	accounts, err := apic.ListAccounts()
	if err != nil {
		return fn.NewE(err)
	}

	var selectedAccount *apiclient.Account

	if len(accounts) == 0 {
		return fn.Error("no accounts found")
	} else if len(accounts) == 1 {
		selectedAccount = &accounts[0]
	} else {
		selectedAccount, err = fzf.FindOne(accounts, func(item apiclient.Account) string {
			return item.Metadata.Name
		}, fzf.WithPrompt("Select account to use >"))
		if err != nil {
			return err
		}
	}

	k, err := k3s.NewClient()
	if err != nil {
		return err
	}
	if err = k.CreateClustersAccounts(selectedAccount.Metadata.Name); err != nil {
		return fn.NewE(err)
	}
	fn.Log("Selected account is ", selectedAccount.Metadata.Name)
	return nil
}
