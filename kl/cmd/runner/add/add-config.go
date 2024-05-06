package add

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

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

var confCmd = &cobra.Command{
	Use:   "config",
	Short: "Add config references to your kl-config",
	Long: `
This command will add config entry references from current environment to your kl-config file.
	`,
	Example: `
  kl add config 		# add config and entry by selecting from list
  kl add config --name <name> 	# add entry by providing config name
  kl add config <name>		# add all entries of config by providing config name
  
  # Customise your mapping to local keys
  kl add config <name> -m <ref_key>=<your_local_key>
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
	name := fn.ParseStringFlag(cmd, "name")
	m := fn.ParseStringFlag(cmd, "map")

	filePath := fn.ParseKlFile(cmd)

	if name == "" && len(args) >= 1 {
		name = args[0]
	}

	klFile, err := client.GetKlFile(filePath)
	if err != nil {
		fn.PrintError(err)
		es := "please run 'kl init' if you are not initialized the file already"
		return fmt.Errorf(es)
	}

	configs, err := server.ListConfigs()
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		return errors.New("no configs created yet on server")
	}

	selectedConfigGroup := server.Config{}

	if name != "" {
		for _, c := range configs {
			if c.Metadata.Name == name {
				selectedConfigGroup = c
				break
			}
		}
		return errors.New("can't find configs with provided name")
	} else {

		selectedGroup, e := fzf.FindOne(
			configs,
			func(item server.Config) string {
				return item.Metadata.Name
			},
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

	if m != "" {
		kk := strings.Split(m, "=")
		if len(kk) != 2 {
			return errors.New("map must be in format of config_key=your_var_key")
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

		return errors.New("config_key not found in selected config")

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
			return err
		}
	}
	// var found bool
	// for i, envVar := range klFile.EnvVars {
	// 	if envVar.Key == selectedConfigKey.Key {
	// 		klFile.EnvVars[i].Value = selectedConfigKey.Value
	// 		found = true
	// 		break
	// 	}
	// }
	// if !found {
	// 	klFile.EnvVars = append(klFile.EnvVars, client.EnvType{
	// 		Key:   selectedConfigKey.Key,
	// 		Value: selectedConfigKey.Value,
	// 	})
	// }

	matchedGroupIndex := -1
	for i, rt := range klFile.Configs {
		if rt.Name == selectedConfigGroup.Metadata.Name {
			matchedGroupIndex = i
			break
		}
	}

	if matchedGroupIndex != -1 {
		matchedKeyIndex := -1

		for i, ret := range klFile.Configs[matchedGroupIndex].Env {
			if ret.RefKey == selectedConfigKey.Key {
				matchedKeyIndex = i
				break
			}
		}

		if matchedKeyIndex == -1 {
			klFile.Configs[matchedGroupIndex].Env = append(klFile.Configs[matchedGroupIndex].Env, client.ResEnvType{
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
		klFile.Configs = append(klFile.Configs, client.ResType{
			Name: selectedConfigGroup.Metadata.Name,
			Env: []client.ResEnvType{
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

	err = client.WriteKLFile(*klFile)
	if err != nil {
		return err
	}

	fn.Log(fmt.Sprintf("added config %s/%s to your kl-file\n", selectedConfigGroup.Metadata.Name, selectedConfigKey.Key))

	return nil
}

func init() {
	confCmd.Flags().StringP("map", "m", "", "config_key=your_var_key")
	confCmd.Flags().StringP("name", "n", "", "config name")
	confCmd.Aliases = append(confCmd.Aliases, "conf")
	fn.WithKlFile(confCmd)
}
