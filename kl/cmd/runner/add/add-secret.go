package add

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"

	"github.com/kloudlite/kl/constants"
	"github.com/spf13/cobra"
)

var addSecretCommand = &cobra.Command{
	Use:   "secret",
	Short: "add secret to your " + constants.CmdName + "-config file by selection from the all the secrets available in selected project",
	Long: `Add env from secret

Using this command you are able to add a environment from the secret present on your project
Examples:
  # add secret
  kl add secret

	# add secret by providing name of secret
	kl add secret --name <name>
	kl add secret <name>

	# add secret by providing your key and refkey
	kl add secret <name> --map [ref_key]=[your_local_key]
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
	m := cmd.Flag("map").Value.String()
	name := cmd.Flag("name").Value.String()

	if name == "" && len(args) >= 1 {
		name = args[0]
	}

	klFile, err := client.GetKlFile(nil)
	if err != nil {
		fn.PrintError(err)
		es := "please run '" + constants.CmdName + " init' if you are not initialized the file already"
		return fmt.Errorf(es)
	}

	secrets, err := server.ListSecrets()
	if err != nil {
		return err
	}

	if len(secrets) == 0 {
		return fmt.Errorf("no secrets created yet on server")
	}

	selectedSecretGroup := server.Secret{}

	if name != "" {
		for _, c := range secrets {
			if c.Metadata.Name == name {
				selectedSecretGroup = c
				break
			}
		}
		return errors.New("can't find secrets with provided name")

	} else {
		selectedGroup, err := fzf.FindOne(
			secrets,
			func(item server.Secret) string {
				return item.Metadata.Name
			},
			fzf.WithPrompt("Select Secret Group >"),
		)
		if err != nil {
			return err
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

	if m != "" {
		kk := strings.Split(m, "=")
		if len(kk) != 2 {
			return errors.New("map must be in format of secret_key=your_var_key")
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

		return errors.New("secret_key not found in selected secret")

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
			return err
		}
	}

	matchedGroupIndex := -1
	for i, rt := range klFile.Secrets {
		if rt.Name == selectedSecretGroup.Metadata.Name {
			matchedGroupIndex = i
			break
		}
	}

	if matchedGroupIndex != -1 {
		matchedKeyIndex := -1

		for i, ret := range klFile.Secrets[matchedGroupIndex].Env {
			if ret.RefKey == selectedSecretKey.Key {
				matchedKeyIndex = i
				break
			}
		}

		if matchedKeyIndex == -1 {
			klFile.Secrets[matchedGroupIndex].Env = append(klFile.Secrets[matchedGroupIndex].Env, client.ResEnvType{
				Key: func() string {
					if m != "" {
						kk := strings.Split(m, "=")
						return kk[1]
					}
					return selectedSecretKey.Key
				}(),
				RefKey: selectedSecretKey.Key,
			})
		}
	} else {
		klFile.Secrets = append(klFile.Secrets, client.ResType{
			Name: selectedSecretGroup.Metadata.Name,
			Env: []client.ResEnvType{
				{
					Key: func() string {
						if m != "" {
							kk := strings.Split(m, "=")
							return kk[1]
						}
						return selectedSecretKey.Key
					}(),
					RefKey: selectedSecretKey.Key,
				},
			},
		})

	}

	err = client.WriteKLFile(*klFile)
	if err != nil {
		fn.PrintError(err)
	}

	fmt.Printf("added secret %s/%s to your %s-file\n", selectedSecretGroup.Metadata.Name, selectedSecretKey.Key, constants.CmdName)
	return nil
}

func init() {
	k := ""
	addSecretCommand.Flags().StringVarP(&k, "map", "", "", "secret_key=your_var_key")

	l := ""
	addSecretCommand.Flags().StringVarP(&l, "name", "", "", "secret name")
}
