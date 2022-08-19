package add

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/constants"
	"kloudlite.io/cmd/internal/lib/server"
)

var addSecretCommand = &cobra.Command{
	Use:   "secret",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		selectSecret()
	},
}

func selectSecret() {

	klFile, err := server.GetKlFile(nil)
	if err != nil {
		common.PrintError(err)
		es := "Please run '" + constants.CMD_NAME + " init' if you are not initialized the file already"
		common.PrintError(fmt.Errorf(es))
		return
	}

	secrets, err := server.GetSecrets()
	if err != nil {
		common.PrintError(err)
		return
	}

	if len(secrets) == 0 {
		es := "No secrets created yet on server"
		common.PrintError(fmt.Errorf(es))
		return
	}

	selectedGroupIndex, err := fuzzyfinder.Find(
		secrets,
		func(i int) string {
			return secrets[i].Name
		},
		fuzzyfinder.WithPromptString("Select Secret Group >"),
	)
	if err != nil {
		common.PrintError(err)
	}

	selectedSecretGroup := secrets[selectedGroupIndex]

	if len(selectedSecretGroup.Entries) == 0 {
		es := fmt.Sprintf("No secrets added yet to %s secret", selectedSecretGroup.Name)
		common.PrintError(fmt.Errorf(es))
		return
	}

	selectedKeyIndex, err := fuzzyfinder.Find(
		selectedSecretGroup.Entries,
		func(i int) string {
			return selectedSecretGroup.Entries[i].Key
		},
		fuzzyfinder.WithPromptString(fmt.Sprintf("Select Key of %s >", selectedSecretGroup.Name)),
	)
	if err != nil {
		common.PrintError(err)
	}

	selectedSecretKey := selectedSecretGroup.Entries[selectedKeyIndex]

	matchedGroupIndex := -1
	for i, rt := range klFile.Secrets {
		if rt.Id == selectedSecretGroup.Id {
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
				Key:    selectedSecretKey.Key,
				RefKey: selectedSecretKey.Key,
			})
		}
	} else {
		klFile.Secrets = append(klFile.Secrets, server.ResType{
			Id:   selectedSecretGroup.Id,
			Name: selectedSecretGroup.Name,
			Env: []server.ResEnvType{
				{
					Key:    selectedSecretKey.Key,
					RefKey: selectedSecretKey.Key,
				},
			},
		})

	}

	err = server.WriteKLFile(*klFile)
	if err != nil {
		common.PrintError(err)
	}

	fmt.Printf("added secret %s/%s to your %s-file\n", selectedSecretGroup.Name, selectedSecretKey.Key, constants.CMD_NAME)
}
