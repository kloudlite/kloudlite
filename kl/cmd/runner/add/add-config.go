package add

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	server2 "github.com/kloudlite/kl/domain/server"
	common_util "github.com/kloudlite/kl/pkg/functions"
	"strings"

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
			common_util.PrintError(err)
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
		common_util.PrintError(err)
		es := "please run '" + constants.CmdName + " init' if you are not initialized the file already"
		return fmt.Errorf(es)
	}

	configs, err := server2.GetConfigs()
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		return errors.New("no configs created yet on server")
	}

	selectedConfigGroup := server2.Config{}

	if name != "" {
		for _, c := range configs {
			if c.Name == name {
				selectedConfigGroup = c
				break
			}
		}
		return errors.New("can't find configs with provided name")
	} else {

		selectedGroupIndex, e := fuzzyfinder.Find(
			configs,
			func(i int) string {
				return configs[i].Name
			},
			fuzzyfinder.WithPromptString("Select Config Group >"),
		)
		if e != nil {
			return e
		}

		selectedConfigGroup = configs[selectedGroupIndex]
	}

	if len(selectedConfigGroup.Entries) == 0 {
		return fmt.Errorf("no configs added yet to %s config", selectedConfigGroup.Name)
	}

	selectedConfigKey := server2.CSEntry{}

	if m != "" {
		kk := strings.Split(m, "=")
		if len(kk) != 2 {
			return errors.New("map must be in format of config_key=your_var_key")
		}

		for _, c := range selectedConfigGroup.Entries {
			if c.Key == kk[0] {
				selectedConfigKey = c
				break
			}
		}

		return errors.New("config_key not found in selected config")

	} else {
		selectedKeyIndex, e := fuzzyfinder.Find(
			selectedConfigGroup.Entries,
			func(i int) string {
				return selectedConfigGroup.Entries[i].Key
			},
			fuzzyfinder.WithPromptString(fmt.Sprintf("Select Key of %s >", selectedConfigGroup.Name)),
		)
		if e != nil {
			return e
		}

		selectedConfigKey = selectedConfigGroup.Entries[selectedKeyIndex]

	}

	matchedGroupIndex := -1
	for i, rt := range klFile.Configs {
		if rt.Name == selectedConfigGroup.Name {
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
			Name: selectedConfigGroup.Name,
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

	fmt.Printf("added config %s/%s to your %s-file\n", selectedConfigGroup.Name, selectedConfigKey.Key, constants.CmdName)

	return nil
}

func init() {
	k := ""
	addConfigCommand.Flags().StringVarP(&k, "map", "", "", "config_key=your_var_key")

	l := ""
	addConfigCommand.Flags().StringVarP(&l, "name", "", "", "config name")
}
