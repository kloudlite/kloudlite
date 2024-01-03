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
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var addConfigCommand = &cobra.Command{
	Use:   "config",
	Short: "add config to your " + constants.CmdName + "-config file by selection from the all the config available in selected project",
	Long: `Add env from managed resource

Using this command you are able to add a environment from the config present on your project
Examples:
  # add config
  kl add config

	# add config by providing name of config
	kl add config --name <name>
	kl add config <name>

	# add config by providing your key and refkey
	kl add config <name> --map [ref_key]=[your_local_key]
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

		selectedGroupIndex, e := fuzzyfinder.Find(
			configs,
			func(i int) string {
				return configs[i].Metadata.Name
			},
			fuzzyfinder.WithPromptString("Select Config Group >"),
		)
		if e != nil {
			return e
		}

		selectedConfigGroup = configs[selectedGroupIndex]
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
				Key: func() string {
					if m != "" {
						kk := strings.Split(m, "=")
						return kk[1]
					}
					return selectedConfigKey.Key
				}(),
				RefKey: selectedConfigKey.Key,
			})
		}
	} else {
		klFile.Configs = append(klFile.Configs, client.ResType{
			Name: selectedConfigGroup.Metadata.Name,
			Env: []client.ResEnvType{
				{
					Key: func() string {
						if m != "" {
							kk := strings.Split(m, "=")
							return kk[1]
						}
						return selectedConfigKey.Key
					}(),
					RefKey: selectedConfigKey.Key,
				},
			},
		})
	}

	err = client.WriteKLFile(*klFile)
	if err != nil {
		return err
	}

	fmt.Printf("added config %s/%s to your %s-file\n", selectedConfigGroup.Metadata.Name, selectedConfigKey.Key, constants.CmdName)

	return nil
}

func init() {
	k := ""
	addConfigCommand.Flags().StringVarP(&k, "map", "", "", "config_key=your_var_key")

	l := ""
	addConfigCommand.Flags().StringVarP(&l, "name", "", "", "config name")
}
