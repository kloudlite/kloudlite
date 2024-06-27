package del

// import (
// 	"fmt"

// 	"github.com/kloudlite/kl/domain/fileclient"
// 	fn "github.com/kloudlite/kl/pkg/functions"
// 	"github.com/kloudlite/kl/pkg/ui/fzf"

// 	"github.com/spf13/cobra"
// )

// var deleteConfigCommand = &cobra.Command{
// 	Use:   "config",
// 	Short: "remove one config environment from your kl-config",
// 	Long: `This command help you to delete environment that that is comming from config

// Examples:
//   # remove config
//   kl del config
// 	`,
// 	Run: func(_ *cobra.Command, _ []string) {
// 		err := removeConfig()
// 		if err != nil {
// 			fn.PrintError(err)
// 			return
// 		}
// 	},
// }

// func removeConfig() error {

// 	klFile, err := fileclient.GetKlFile("")
// 	if err != nil {
// 		fn.PrintError(err)
// 		es := "please run 'kl init' if you are not initialized the file already"
// 		fn.PrintError(fmt.Errorf(es))
// 	}

// 	if len(klFile.Configs) == 0 {
// 		es := "no configs added yet in your file"
// 		return fmt.Errorf(es)
// 	}

// 	selectedConfig, err := fzf.FindOne(
// 		klFile.Configs,
// 		func(item fileclient.ResType) string {
// 			return item.Name
// 		},
// 		fzf.WithPrompt("Select Config Group >"),
// 	)

// 	if err != nil {
// 		return functions.NewE(err)
// 	}

// 	if len(selectedConfig.Env) == 1 {
// 		newConfigs := make([]fileclient.ResType, 0)
// 		for _, rt := range klFile.Configs {
// 			if rt.Name == selectedConfig.Name {
// 				continue
// 			}
// 			newConfigs = append(newConfigs, rt)
// 		}

// 		klFile.Configs = newConfigs

// 		fn.Logf("removed config %s form your kl-file\n", selectedConfig.Name)

// 	} else {

// 		selectedKeyValues, e := fzf.FindOne(
// 			selectedConfig.Env,
// 			func(item fileclient.ResEnvType) string {
// 				return fmt.Sprintf("%s, %s", item.Key, item.RefKey)
// 			},
// 			fzf.WithPrompt("Select Config Key >"),
// 		)

// 		if e != nil {
// 			return e
// 		}

// 		newEnvs := make([]fileclient.ResEnvType, 0)
// 		for _, ret := range selectedConfig.Env {
// 			if ret.Name == selectedKeyValues.Name {
// 				continue
// 			}
// 			newEnvs = append(newEnvs, ret)
// 		}

// 		selectedConfig.Env = newEnvs

// 		fn.Logf("removed key %s/%s form your kl-file\n", selectedConfig.Name, selectedConfig.Name)
// 	}

// 	err = fileclient.WriteKLFile(*klFile)

// 	return functions.NewE(err)
// }
