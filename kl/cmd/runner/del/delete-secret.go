package del

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	common_util "github.com/kloudlite/kl/pkg/functions"

	"github.com/kloudlite/kl/constants"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var deleteSecretCommand = &cobra.Command{
	Use:   "secret",
	Short: "remove one secret environment from your " + constants.CmdName + "-config",
	Long: `This command help you to delete environment that that is comming from secret

Examples:
  # remov secret
  kl del secret
`,
	Run: func(_ *cobra.Command, _ []string) {
		err := removeSecret()

		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}

func removeSecret() error {

	klFile, err := client.GetKlFile(nil)
	if err != nil {
		common_util.PrintError(err)
		es := "please run '" + constants.CmdName + " init' if you are not initialized the file already"
		return fmt.Errorf(es)
	}

	if len(klFile.Secrets) == 0 {
		es := "no secrets added yet in your file"
		return fmt.Errorf(es)
	}

	selectedSecretIndex, err := fuzzyfinder.Find(
		klFile.Secrets,
		func(i int) string {
			return klFile.Secrets[i].Name
		},
		fuzzyfinder.WithPromptString("Select Secret Group >"),
	)

	if err != nil {
		return err
	}

	selectedSecret := klFile.Secrets[selectedSecretIndex]

	if len(selectedSecret.Env) == 1 {
		newSecrets := make([]client.ResType, 0)
		for i, rt := range klFile.Secrets {
			if i == selectedSecretIndex {
				continue
			}
			newSecrets = append(newSecrets, rt)
		}

		klFile.Secrets = newSecrets

		fmt.Printf("removed secret %s form your %s-file\n", selectedSecret.Name, constants.CmdName)

	} else {

		selectedKeyIndex, e := fuzzyfinder.Find(
			selectedSecret.Env,
			func(i int) string {
				return fmt.Sprintf("%s, %s", selectedSecret.Env[i].Key, selectedSecret.Env[i].RefKey)
			},
			fuzzyfinder.WithPromptString("Select Secret Key >"),
		)

		if e != nil {
			return e
		}

		newEnvs := make([]client.ResEnvType, 0)
		for i, ret := range selectedSecret.Env {
			if i == selectedKeyIndex {
				continue
			}
			newEnvs = append(newEnvs, ret)
		}

		klFile.Secrets[selectedSecretIndex].Env = newEnvs

		fmt.Printf("removed key %s/%s form your %s-file\n", selectedSecret.Name, selectedSecret.Name, constants.CmdName)
	}

	err = client.WriteKLFile(*klFile)

	return err
}
