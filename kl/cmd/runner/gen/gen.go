package gen

import (
	"fmt"

	"github.com/kloudlite/kl/constants"
	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var MountCommand = &cobra.Command{
	Use:   "gen",
	Short: "add mount to your " + constants.CmdName + "-config file by selection from the all the [ config | secret ] available selected project",
	Long: `Add mount
This command help you to add generated config so you can get your config or secret downloaded and dumped in a file.

Examples:
  # add config to download at the time of load
  kl gen --configName=<config_name>

  # add secret to download at the time of load
  kl gen --secretName=<secret_name>
`,
	Run: func(cmd *cobra.Command, args []string) {
		klFile, err := server.GetKlFile(nil)
		if err != nil {
			common_util.PrintError(err)
			es := "please run '" + constants.CmdName + " init' if you are not initialized the file already"
			common_util.PrintError(fmt.Errorf(es))
			return
		}
		path := ""

		if len(args) > 0 {
			path = args[0]
		}

		err = selectConfigMount(path, *klFile, cmd)
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}

func selectConfigMount(path string, klFile server.KLFileType, cmd *cobra.Command) error {

	c := cmd.Flag("config").Value.String()
	s := cmd.Flag("secret").Value.String()

	cOrs := ""

	if c != "" || s != "" {

		if c != "" {
			cOrs = "config"
		} else {
			cOrs = "secret"
		}

	} else {
		csName := []string{"config", "secret"}
		cOrsIndex, err := fuzzyfinder.Find(
			csName,
			func(i int) string {
				return csName[i]
			},
			fuzzyfinder.WithPromptString("Mount from Config/Secret >"),
		)
		if err != nil {
			return err
		}

		cOrs = csName[cOrsIndex]
	}

	items := make([]server.ConfigORSecret, 0)
	if cOrs == "config" {
		configs, e := server.GetConfigs()

		if e != nil {
			return e
		}

		for _, c := range configs {
			items = append(items, server.ConfigORSecret{
				Name: c.Name,
			})
		}

	} else {
		secrets, e := server.GetSecrets()

		if e != nil {
			return e
		}

		for _, c := range secrets {
			items = append(items, server.ConfigORSecret{
				Entries: c.Entries,
				Name:    c.Name,
			})
		}
	}

	if len(items) == 0 {
		return fmt.Errorf("no %ss created yet on server", cOrs)
	}

	selectedItem := server.ConfigORSecret{}

	if c != "" || s != "" {
		csId := func() string {
			if c != "" {
				return c
			}
			return s
		}()

		for _, co := range items {
			if co.Name == csId {
				selectedItem = co
				break
			}
		}

		return fmt.Errorf("provided %s name not found", cOrs)
	} else {
		selectedItemIndex, err := fuzzyfinder.Find(
			items,
			func(i int) string {
				return items[i].Name
			},
			fuzzyfinder.WithPromptString(fmt.Sprintf("Select %s >", cOrs)),
		)

		if err != nil {
			common_util.PrintError(err)
		}

		selectedItem = items[selectedItemIndex]
	}

	matchedIndex := -1
	for i, fe := range klFile.FileMount.Mounts {
		if fe.Path == path {
			matchedIndex = i
		}
	}

	if matchedIndex == -1 {
		klFile.FileMount.Mounts = append(klFile.FileMount.Mounts, server.FileEntry{
			Type: cOrs,
			Path: path,
			Name: selectedItem.Name,
		})
	} else {
		klFile.FileMount.Mounts[matchedIndex] = server.FileEntry{
			Type: cOrs,
			Path: path,
			Name: selectedItem.Name,
		}
	}

	err := server.WriteKLFile(klFile)

	if err != nil {
		return err
	}

	fmt.Println("Mount added to config")
	return nil
}

func init() {
	MountCommand.Flags().StringP("config", "", "", "config name")
	MountCommand.Flags().StringP("secret", "", "", "secret name")
}
