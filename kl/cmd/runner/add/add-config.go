package add

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var confCmd = &cobra.Command{
	Use:   "config [name]",
	Short: "add config references to your kl-config",
	Long: `
This command will add config entry references from current environment to your kl-config file.
	`,
	Example: `
  kl add config 		# add config and entry by selecting from list
  kl add config [name] 		# add all entries of config by providing config name
	`,
	Run: func(cmd *cobra.Command, args []string) {
		err := selectAndAddConfig(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func selectAndAddConfig(cmd *cobra.Command, args []string) error {
	fc, err := fileclient.New()
	if err != nil {
		return fn.NewE(err)
	}

	apic, err := apiclient.New()
	if err != nil {
		return fn.NewE(err)
	}

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

	configs, err := apic.ListConfigs(currentAccount, currentEnv.Name)
	if err != nil {
		return fn.NewE(err)
	}

	if len(configs) == 0 {
		return fn.Error("no configs created yet on server")
	}

	selectedConfigGroup := apiclient.Config{}

	if name != "" {
		for _, c := range configs {
			if c.Metadata.Name == name {
				selectedConfigGroup = c
				break
			}
		}
		return fn.Error("can't find configs with provided name")
	} else {

		selectedGroup, e := fzf.FindOne(
			configs,
			func(item apiclient.Config) string { return item.Metadata.Name },
			fzf.WithPrompt("Select Config Group >"),
		)
		if e != nil {
			return e
		}

		selectedConfigGroup = *selectedGroup
	}

	if len(selectedConfigGroup.Data) == 0 {
		return fmt.Errorf("no configs added yet to %s config", selectedConfigGroup.Metadata.Name)
	}

	type KV struct {
		Key   string
		Value string
	}

	selectedConfigKey := &KV{}

	m := ""
	if m != "" {
		kk := strings.Split(m, "=")
		if len(kk) != 2 {
			return fn.Error("map must be in format of config_key=your_var_key")
		}

		for k, v := range selectedConfigGroup.Data {
			if k == kk[0] {
				selectedConfigKey = &KV{
					Key:   k,
					Value: v,
				}
				break
			}
		}

		return fn.Error("config_key not found in selected config")

	} else {
		selectedConfigKey, err = fzf.FindOne(
			func() []KV {
				var kvs []KV

				for k, v := range selectedConfigGroup.Data {
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
			fzf.WithPrompt(fmt.Sprintf("Select Key of %s >", selectedConfigGroup.Metadata.Name)),
		)
		if err != nil {
			return fn.NewE(err)
		}
	}

	matchedGroupIndex := -1
	for i, rt := range klFile.EnvVars.GetConfigs() {
		if rt.Name == selectedConfigGroup.Metadata.Name {
			matchedGroupIndex = i
			break
		}
	}

	currConfigs := klFile.EnvVars.GetConfigs()

	//for i, ret := range currConfigs {
	//	fmt.Println(ret.Name, selectedConfigGroup.Metadata.Name)
	//	if ret.Name == selectedConfigGroup.Metadata.Name {
	//		for j, rt := range currConfigs[i].Env {
	//			fmt.Println(rt.RefKey, selectedConfigKey.Key, j)
	//			if rt.RefKey == selectedConfigKey.Key {
	//				//if len(currConfigs) >= 1 {
	//				//	currConfigs = []fileclient.ResType{}
	//				//	matchedGroupIndex = -1
	//				//	break
	//				//}
	//				//currConfigs = append(currConfigs[:i], currConfigs[i+1:]...)
	//				klFile.EnvVars = append(klFile.EnvVars[:i], klFile.EnvVars[i+1:]...)
	//			}
	//		}
	//	}
	//	err := fileclient.WriteKLFile(*klFile)
	//	if err != nil {
	//		return functions.NewE(err)
	//	}
	//	klFile, err = fileclient.GetKlFile("")
	//	if err != nil {
	//		return functions.NewE(err)
	//	}
	//}
	//fmt.Println(currConfigs, matchedGroupIndex)

	if matchedGroupIndex != -1 {
		matchedKeyIndex := -1

		for i, ret := range currConfigs[matchedGroupIndex].Env {
			if ret.RefKey == selectedConfigKey.Key {
				matchedKeyIndex = i
				break
			}
		}
		if matchedKeyIndex == -1 {
			currConfigs[matchedGroupIndex].Env = append(currConfigs[matchedGroupIndex].Env, fileclient.ResEnvType{
				Key: RenameKey(func() string {
					if m != "" {
						kk := strings.Split(m, "=")
						return kk[1]
					}
					return selectedConfigKey.Key
				}()),
				RefKey: selectedConfigKey.Key,
			})
		}
	} else {
		currConfigs = append(currConfigs, fileclient.ResType{
			Name: selectedConfigGroup.Metadata.Name,
			Env: []fileclient.ResEnvType{
				{
					Key: RenameKey(func() string {
						if m != "" {
							kk := strings.Split(m, "=")
							return kk[1]
						}
						return selectedConfigKey.Key
					}()),
					RefKey: selectedConfigKey.Key,
				},
			},
		})
	}

	klFile.EnvVars.AddResTypes(currConfigs, fileclient.Res_config)

	err = fc.WriteKLFile(*klFile)
	if err != nil {
		return fn.NewE(err)
	}

	fn.Log(fmt.Sprintf("added config %s/%s to your kl-file\n", selectedConfigGroup.Metadata.Name, selectedConfigKey.Key))

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

func RenameKey(key string) string {
	regexPattern := `[^a-zA-Z0-9]`

	regexpCompiled, err := regexp.Compile(regexPattern)
	if err != nil {
		fn.Log(text.Yellow(fmt.Sprintf("[#] error compiling regex pattern: %s", regexPattern)))
		return key
	}

	resultString := regexpCompiled.ReplaceAllString(key, "_")

	return strings.ToUpper(resultString)
}

func init() {
	// confCmd.Flags().StringP("map", "m", "", "config_key=your_var_key")
	confCmd.Flags().StringP("name", "n", "", "config name")
	confCmd.Aliases = append(confCmd.Aliases, "conf")
	fn.WithKlFile(confCmd)
}
