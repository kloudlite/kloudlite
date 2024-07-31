package add

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/kloudlite/kl/pkg/ui/fzf"

	"github.com/spf13/cobra"
)

var mresCmd = &cobra.Command{
	Use:   "mres [name]",
	Short: "add managed resource references to your kl-config",
	Long: `
This command will add secret entry of managed resource references from current environement to your kl-config file.
`,
	Example: ` 
  kl add mres # add mres secret entry to your kl-config as env var
  kl add  mres [name] # add specific mres secret entry to your kl-config as env var by providing mres name
`,
	Run: func(cmd *cobra.Command, args []string) {
		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		if err := AddMres(apic, fc, cmd, args); err != nil {
			fn.PrintError(err)
			return
		}

	},
}

func AddMres(apic apiclient.ApiClient, fc fileclient.FileClient, cmd *cobra.Command, args []string) error {

	filePath := fn.ParseKlFile(cmd)
	kt, err := fc.GetKlFile(filePath)
	if err != nil {
		return fn.NewE(err)
	}

	//TODO: add changes to the klbox-hash file
	// mresName := fn.ParseStringFlag(cmd, "resource")

	mres, err := selectMres(apic, fc)

	if err != nil {
		return fn.NewE(err)
	}

	mresKey, err := selectMresKey(apic, fc, mres.SecretRefName.Name)

	if err != nil {
		return fn.NewE(err)
	}

	currMreses := kt.EnvVars.GetMreses()

	if currMreses == nil {
		currMreses = []fileclient.ResType{
			{
				Name: mres.SecretRefName.Name,
				Env: []fileclient.ResEnvType{
					{
						Key:    RenameKey(fmt.Sprintf("%s_%s", mres.SecretRefName.Name, *mresKey)),
						RefKey: *mresKey,
					},
				},
			},
		}
	}

	if currMreses != nil {
		matchedMres := false
		for i, rt := range currMreses {
			if rt.Name == mres.SecretRefName.Name {
				currMreses[i].Env = append(currMreses[i].Env, fileclient.ResEnvType{
					Key:    RenameKey(fmt.Sprintf("%s_%s", mres.SecretRefName.Name, *mresKey)),
					RefKey: *mresKey,
				})
				matchedMres = true
				break
			}
		}

		if !matchedMres {
			currMreses = append(currMreses, fileclient.ResType{
				Name: mres.SecretRefName.Name,
				Env: []fileclient.ResEnvType{
					{
						Key:    RenameKey(fmt.Sprintf("%s_%s", mres.SecretRefName.Name, *mresKey)),
						RefKey: *mresKey,
					},
				},
			})
		}
	}

	kt.EnvVars.AddResTypes(currMreses, fileclient.Res_mres)
	if err := fc.WriteKLFile(*kt); err != nil {
		return fn.NewE(err)
	}

	fn.Log(fmt.Sprintf("added mres %s/%s to your kl-file", mres.SecretRefName.Name, *mresKey))

	wpath, err := os.Getwd()
	if err != nil {
		return fn.NewE(err)
	}

	if err := hashctrl.SyncBoxHash(apic, fc, wpath); err != nil {
		return fn.NewE(err)
	}

	c, err := boxpkg.NewClient(cmd, args)
	if err != nil {
		return err
	}

	if err := c.ConfirmBoxRestart(); err != nil {
		return fn.NewE(err)
	}

	return nil
}

func selectMres(apic apiclient.ApiClient, fc fileclient.FileClient) (*apiclient.Mres, error) {
	currentEnv, err := apic.EnsureEnv()
	if err != nil {
		return nil, fn.NewE(err)
	}
	currentAccount, err := fc.CurrentAccountName()
	if err != nil {
		return nil, fn.NewE(err)
	}
	m, err := apic.ListMreses(currentAccount, currentEnv.Name)
	if err != nil {
		return nil, fn.NewE(err)
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("no managed resources created yet on server")
	}

	mres, err := fzf.FindOne(m, func(item apiclient.Mres) string {
		return item.DisplayName
	}, fzf.WithPrompt("Select managed resource >"))

	return mres, err
}

func init() {
	fn.WithKlFile(mresCmd)
}

func selectMresKey(apic apiclient.ApiClient, fc fileclient.FileClient, secretName string) (*string, error) {
	selectedAccount, err := fc.CurrentAccountName()
	if err != nil {
		return nil, fn.NewE(err)
	}
	secret, err := apic.GetSecret(selectedAccount, secretName)
	if err != nil {
		return nil, fn.NewE(err)
	}

	keys := []string{}
	for k := range secret.StringData {
		keys = append(keys, k)
	}

	key, err := fzf.FindOne(keys, func(item string) string {
		return item
	}, fzf.WithPrompt("Select key >"))

	return key, err
}
