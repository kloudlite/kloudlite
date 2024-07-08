package add

import (
	"fmt"
	"os"
	"strings"

	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"

	"github.com/spf13/cobra"
)

var secCmd = &cobra.Command{
	Use:   "secret [name]",
	Short: "add secret references to your kl-config",
	Long:  `This command will add secret entry references from current environement to your kl-config file.`,
	Example: `
  kl add secret 		# add secret and entry by selecting from list (default)
  kl add secret [name] 	# add entry by providing secret name
	`,
	Run: func(cmd *cobra.Command, args []string) {
		err := selectAndAddSecret(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func selectAndAddSecret(cmd *cobra.Command, args []string) error {
	fc, err := fileclient.New()
	if err != nil {
		return fn.NewE(err)
	}

	apic, err := apiclient.New()
	if err != nil {
		return fn.NewE(err)
	}

	//TODO: add changes to the klbox-hash file
	// m := fn.ParseStringFlag(cmd, "map")
	filePath := fn.ParseKlFile(cmd)

	name := ""
	if len(args) >= 1 {
		name = args[0]
	}

	klFile, err := fc.GetKlFile(filePath)
	if err != nil {
		return fn.NewE(err)
	}

	currentAccount, err := fc.CurrentAccountName()
	if err != nil {
		return fn.NewE(err)
	}
	currentEnv, err := fc.CurrentEnv()
	if err != nil {
		return fn.NewE(err)
	}

	secrets, err := apic.ListSecrets(currentAccount, currentEnv.Name)
	if err != nil {
		return functions.NewE(err)
	}

	if len(secrets) == 0 {
		return fmt.Errorf("no secrets created yet on server")
	}

	selectedSecretGroup := apiclient.Secret{}

	if name != "" {
		for _, c := range secrets {
			if c.Metadata.Name == name {
				selectedSecretGroup = c
				break
			}
		}
		return functions.Error("can't find secrets with provided name")

	} else {
		selectedGroup, err := fzf.FindOne(
			secrets,
			func(item apiclient.Secret) string {
				return item.Metadata.Name
			},
			fzf.WithPrompt("Select Secret Group >"),
		)
		if err != nil {
			return functions.NewE(err)
		}

		selectedSecretGroup = *selectedGroup
	}

	if len(selectedSecretGroup.StringData) == 0 {
		return fmt.Errorf("no secrets added yet to %s secret", selectedSecretGroup.Metadata.Name)
	}

	type KV struct {
		Key   string
		Value string
	}

	selectedSecretKey := &KV{}

	m := ""

	if m != "" {
		kk := strings.Split(m, "=")
		if len(kk) != 2 {
			return functions.Error("map must be in format of secret_key=your_var_key")
		}

		for k, v := range selectedSecretGroup.StringData {
			if k == kk[0] {
				selectedSecretKey = &KV{
					Key:   k,
					Value: v,
				}
				break
			}
		}

		return functions.Error("secret_key not found in selected secret")

	} else {
		selectedSecretKey, err = fzf.FindOne(
			func() []KV {
				var kvs []KV

				for k, v := range selectedSecretGroup.StringData {
					kvs = append(kvs, KV{
						Key:   k,
						Value: v,
					})
				}

				return kvs
			}(),
			func(val KV) string {
				return val.Key
			},
			fzf.WithPrompt(fmt.Sprintf("Select Key of %s >", selectedSecretGroup.Metadata.Name)),
		)
		if err != nil {
			return functions.NewE(err)
		}
	}

	currSecs := klFile.EnvVars.GetSecrets()

	matchedGroupIndex := -1
	for i, rt := range currSecs {
		if rt.Name == selectedSecretGroup.Metadata.Name {
			matchedGroupIndex = i
			break
		}
	}

	if matchedGroupIndex != -1 {
		matchedKeyIndex := -1

		for i, ret := range currSecs[matchedGroupIndex].Env {
			if ret.RefKey == selectedSecretKey.Key {
				matchedKeyIndex = i
				break
			}
		}

		if matchedKeyIndex == -1 {
			currSecs[matchedGroupIndex].Env = append(currSecs[matchedGroupIndex].Env, fileclient.ResEnvType{
				Key: RenameKey(func() string {
					if m != "" {
						kk := strings.Split(m, "=")
						return kk[1]
					}
					return selectedSecretKey.Key
				}()),
				RefKey: selectedSecretKey.Key,
			})
		}
	} else {
		currSecs = append(currSecs, fileclient.ResType{
			Name: selectedSecretGroup.Metadata.Name,
			Env: []fileclient.ResEnvType{
				{
					Key: RenameKey(func() string {
						if m != "" {
							kk := strings.Split(m, "=")
							return kk[1]
						}
						return selectedSecretKey.Key
					}()),
					RefKey: selectedSecretKey.Key,
				},
			},
		})

	}

	klFile.EnvVars.AddResTypes(currSecs, fileclient.Res_secret)
	err = fc.WriteKLFile(*klFile)
	if err != nil {
		return functions.NewE(err)
	}

	fn.Log(fmt.Sprintf("added secret %s/%s to your kl-file\n", selectedSecretGroup.Metadata.Name, selectedSecretKey.Key))

	wpath, err := os.Getwd()
	if err != nil {
		return functions.NewE(err)
	}

	if err := hashctrl.SyncBoxHash(apic, fc, wpath); err != nil {
		return functions.NewE(err)
	}

	c, err := boxpkg.NewClient(cmd, args)
	if err != nil {
		return functions.NewE(err)
	}

	if err := c.ConfirmBoxRestart(); err != nil {
		return functions.NewE(err)
	}

	//if err := apiclient.SyncDevboxJsonFile(); err != nil {
	//	return functions.NewE(err)
	//}
	//
	//if err := fileclient.SyncDevboxShellEnvFile(cmd); err != nil {
	//	return functions.NewE(err)
	//}
	return nil
}

func init() {
	// secCmd.Flags().StringP("map", "m", "", "secret_key=your_var_key")

	secCmd.Aliases = append(secCmd.Aliases, "sec")
	fn.WithKlFile(secCmd)
}
