package add

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kloudlite/kl/constants"
	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/ktr0731/go-fuzzyfinder"
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
			common_util.PrintError(err)
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

	klFile, err := server.GetKlFile(nil)
	if err != nil {
		common_util.PrintError(err)
		es := "please run '" + constants.CmdName + " init' if you are not initialized the file already"
		return fmt.Errorf(es)
	}

	secrets, err := server.GetSecrets()
	if err != nil {
		return err
	}

	if len(secrets) == 0 {
		return fmt.Errorf("no secrets created yet on server")
	}

	selectedSecretGroup := server.Secret{}

	if name != "" {
		for _, c := range secrets {
			if c.Name == name {
				selectedSecretGroup = c
				break
			}
		}
		return errors.New("can't find secrets with provided name")

	} else {
		selectedGroupIndex, err := fuzzyfinder.Find(
			secrets,
			func(i int) string {
				return secrets[i].Name
			},
			fuzzyfinder.WithPromptString("Select Secret Group >"),
		)
		if err != nil {
			return err
		}

		selectedSecretGroup = secrets[selectedGroupIndex]
	}

	if len(selectedSecretGroup.Entries) == 0 {
		return fmt.Errorf("no secrets added yet to %s secret", selectedSecretGroup.Name)
	}

	selectedSecretKey := server.CSEntry{}

	if m != "" {
		kk := strings.Split(m, "=")
		if len(kk) != 2 {
			return errors.New("map must be in format of secret_key=your_var_key")
		}

		for _, c := range selectedSecretGroup.Entries {
			if c.Key == kk[0] {
				selectedSecretKey = c
				break
			}
		}

		return errors.New("secret_key not found in selected secret")

	} else {
		selectedKeyIndex, e := fuzzyfinder.Find(
			selectedSecretGroup.Entries,
			func(i int) string {
				return selectedSecretGroup.Entries[i].Key
			},
			fuzzyfinder.WithPromptString(fmt.Sprintf("Select Key of %s >", selectedSecretGroup.Name)),
		)
		if e != nil {
			return e
		}

		selectedSecretKey = selectedSecretGroup.Entries[selectedKeyIndex]
	}

	matchedGroupIndex := -1
	for i, rt := range klFile.Secrets {
		if rt.Name == selectedSecretGroup.Name {
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
			klFile.Secrets[matchedGroupIndex].Env = append(klFile.Secrets[matchedGroupIndex].Env, server.ResEnvType{
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
		klFile.Secrets = append(klFile.Secrets, server.ResType{
			Name: selectedSecretGroup.Name,
			Env: []server.ResEnvType{
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

	err = server.WriteKLFile(*klFile)
	if err != nil {
		common_util.PrintError(err)
	}

	fmt.Printf("added secret %s/%s to your %s-file\n", selectedSecretGroup.Name, selectedSecretKey.Key, constants.CmdName)
	return nil
}

func init() {
	k := ""
	addSecretCommand.Flags().StringVarP(&k, "map", "", "", "secret_key=your_var_key")

	l := ""
	addSecretCommand.Flags().StringVarP(&l, "name", "", "", "secret name")
}
