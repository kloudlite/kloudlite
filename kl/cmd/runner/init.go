package runner

import (
	"errors"
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/domain/apiclient"
	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "initialize a kl-config file",
	Long:  `use this command to initialize a kl-config file`,
	Run: func(_ *cobra.Command, _ []string) {
		if fileclient.InsideBox() {
			fn.PrintError(functions.Error("cannot re-initialize workspace in dev box"))
			return
		}
		_, err := fileclient.GetKlFile("")
		if err == nil {
			fn.Printf(text.Yellow("Workspace is already initilized. Do you want to override? (y/N): "))
			if !fn.Confirm("Y", "N") {
				return
			}
		} else if !errors.Is(err, confighandler.ErrKlFileNotExists) {
			fn.PrintError(err)
			return
		}

		selectedAccount, err := selectAccount()
		if err != nil {
			fn.PrintError(err)
			return
		} else {
			if selectedEnv, err := selectEnv(*selectedAccount); err != nil {
				fn.PrintError(err)
			} else {
				newKlFile := fileclient.KLFileType{
					AccountName: *selectedAccount,
					DefaultEnv:  *selectedEnv,
					Version:     "v1",
					Packages:    []string{"neovim", "git"},
				}
				if err := fileclient.WriteKLFile(newKlFile); err != nil {
					fn.PrintError(err)
				} else {
					fn.Printf(text.Green("Workspace initialized successfully.\n"))
				}
			}
		}

		dir, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := hashctrl.SyncBoxHash(dir); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func selectAccount() (*string, error) {
	if accounts, err := apiclient.ListAccounts(); err == nil {
		if selectedAccount, err := fzf.FindOne(
			accounts,
			func(account apiclient.Account) string {
				return account.Metadata.Name + " #" + account.Metadata.Name
			},
			fzf.WithPrompt("select kloudlite team > "),
		); err != nil {
			return nil, functions.NewE(err)
		} else {
			return &selectedAccount.Metadata.Name, nil
		}
	} else {
		return nil, functions.NewE(err)
	}
}

func selectEnv(accountName string) (*string, error) {
	if envs, err := apiclient.ListEnvs([]fn.Option{
		fn.MakeOption("accountName", accountName),
	}...); err == nil {
		if selectedEnv, err := fzf.FindOne(
			envs,
			func(env apiclient.Env) string {
				return env.Metadata.Name + " #" + env.Metadata.Name
			},
			fzf.WithPrompt("select environment > "),
		); err != nil {
			return nil, functions.NewE(err)
		} else {
			return &selectedEnv.Metadata.Name, nil
		}
	} else {
		return nil, functions.NewE(err)
	}
}

func init() {
	InitCommand.Flags().StringP("account", "a", "", "account name")
	InitCommand.Flags().StringP("file", "f", "", "file name")
}
