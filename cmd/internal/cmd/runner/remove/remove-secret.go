package remove

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/constants"
	"kloudlite.io/cmd/internal/lib/server"
)

var removeSecretCommand = &cobra.Command{
	Use:   "secret",
	Short: "remove one secret environment from your " + constants.CMD_NAME + "-config",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		removeSecret()
	},
}

func removeSecret() {

	klFile, err := server.GetKlFile(nil)
	if err != nil {
		common.PrintError(err)
		es := "Please run '" + constants.CMD_NAME + " init' if you are not initialized the file already"
		common.PrintError(fmt.Errorf(es))
		return
	}

	if len(klFile.Secrets) == 0 {
		es := "No secrets added yet in your file"
		common.PrintError(fmt.Errorf(es))
		return
	}

	selectedSecretIndex, err := fuzzyfinder.Find(
		klFile.Secrets,
		func(i int) string {
			return fmt.Sprintf("%s, %s", klFile.Secrets[i].Name, klFile.Secrets[i].Id)
		},
		fuzzyfinder.WithPromptString("Select Secret Group >"),
	)

	if err != nil {
		common.PrintError(err)
		return
	}

	selectedSecret := klFile.Secrets[selectedSecretIndex]

	if len(selectedSecret.Env) == 1 {
		newSecrets := make([]server.ResType, 0)
		for i, rt := range klFile.Secrets {
			if i == selectedSecretIndex {
				continue
			}
			newSecrets = append(newSecrets, rt)
		}

		klFile.Secrets = newSecrets

		fmt.Printf("removed secret %s form your %s-file\n", selectedSecret.Name, constants.CMD_NAME)

	} else {

		selectedKeyIndex, e := fuzzyfinder.Find(
			selectedSecret.Env,
			func(i int) string {
				return fmt.Sprintf("%s, %s", selectedSecret.Env[i].Key, selectedSecret.Env[i].RefKey)
			},
			fuzzyfinder.WithPromptString("Select Secret Key >"),
		)

		if e != nil {
			common.PrintError(e)
			return
		}

		newEnvs := make([]server.ResEnvType, 0)
		for i, ret := range selectedSecret.Env {
			if i == selectedKeyIndex {
				continue
			}
			newEnvs = append(newEnvs, ret)
		}

		klFile.Secrets[selectedSecretIndex].Env = newEnvs

		fmt.Printf("removed key %s/%s form your %s-file\n", selectedSecret.Name, selectedSecret.Name, constants.CMD_NAME)
	}

	err = server.WriteKLFile(*klFile)
	if err != nil {
		common.PrintError(err)
	}

}
