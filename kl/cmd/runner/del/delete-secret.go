package del

// import (
// 	"fmt"

// 	"github.com/kloudlite/kl/domain/fileclient"
// 	fn "github.com/kloudlite/kl/pkg/functions"
// 	"github.com/kloudlite/kl/pkg/ui/fzf"

// 	"github.com/spf13/cobra"
// )

// var deleteSecretCommand = &cobra.Command{
// 	Use:   "secret",
// 	Short: "remove one secret environment from your kl-config",
// 	Long: `This command help you to delete environment that that is comming from secret

// Examples:
//   # remov secret
//   kl del secret
// `,
// 	Run: func(_ *cobra.Command, _ []string) {
// 		err := removeSecret()

// 		if err != nil {
// 			fn.PrintError(err)
// 			return
// 		}
// 	},
// }

// func removeSecret() error {

// 	klFile, err := fileclient.GetKlFile("")
// 	if err != nil {
// 		fn.PrintError(err)
// 		es := "please run 'kl init' if you are not initialized the file already"
// 		return fn.Errorf(es)
// 	}

// 	if len(klFile.Secrets) == 0 {
// 		es := "no secrets added yet in your file"
// 		return fn.Errorf(es)
// 	}

// 	selectedSecret, err := fzf.FindOne(
// 		klFile.Secrets,
// 		func(item fileclient.ResType) string {
// 			return item.Name
// 		},
// 		fzf.WithPrompt("Select Secret Group >"),
// 	)

// 	if err != nil {
// 		return functions.NewE(err)
// 	}

// 	if len(selectedSecret.Env) == 1 {
// 		newSecrets := make([]fileclient.ResType, 0)
// 		for _, rt := range klFile.Secrets {
// 			if rt.Name == selectedSecret.Name {
// 				continue
// 			}
// 			newSecrets = append(newSecrets, rt)
// 		}

// 		klFile.Secrets = newSecrets

// 		fn.Logf("removed secret %s form your kl-file\n", selectedSecret.Name)

// 	} else {

// 		selectedKeyVal, e := fzf.FindOne(
// 			selectedSecret.Env,
// 			func(item fileclient.ResEnvType) string {
// 				return fmt.Sprintf("%s, %s", item.Key, item.RefKey)
// 			},
// 			fzf.WithPrompt("Select Secret Key >"),
// 		)

// 		if e != nil {
// 			return e
// 		}

// 		newEnvs := make([]fileclient.ResEnvType, 0)
// 		for _, ret := range selectedSecret.Env {
// 			if ret.Name == selectedKeyVal.Name {
// 				continue
// 			}
// 			newEnvs = append(newEnvs, ret)
// 		}

// 		selectedSecret.Env = newEnvs

// 		fn.Logf("removed key %s/%s form your kl-file\n", selectedSecret.Name, selectedSecret.Name)
// 	}

// 	err = fileclient.WriteKLFile(*klFile)

// 	return functions.NewE(err)
// }
