package gen

import (
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"

	"github.com/spf13/cobra"
)

var MountCommand = &cobra.Command{
	Use:   "gen",
	Short: "add mount to your kl-config file by selection from the all the [ config | secret ] available selected project",
	Long: `Add mount
This command help you to add generated config so you can get your config or secret downloaded and dumped in a file.

Examples:
  # add config to download at the time of load
  kl gen --configName=<config_name>

  # add secret to download at the time of load
  kl gen --secretName=<secret_name>
`,
	Run: func(cmd *cobra.Command, args []string) {
		klFile, err := client.GetKlFile(nil)
		if err != nil {
			fn.PrintError(err)
			es := "please run 'kl init' if you are not initialized the file already"
			fn.PrintError(fmt.Errorf(es))
			return
		}
		path := ""

		if len(args) > 0 {
			path = args[0]
		}

		err = selectConfigMount(path, *klFile, cmd)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func selectConfigMount(path string, klFile client.KLFileType, cmd *cobra.Command) error {

	c := cmd.Flag("config").Value.String()
	s := cmd.Flag("secret").Value.String()

	var cOrs client.CSType
	cOrs = ""

	if c != "" || s != "" {

		if c != "" {
			cOrs = client.ConfigType
		} else {
			cOrs = client.SecretType
		}

	} else {
		csName := []client.CSType{client.ConfigType, client.SecretType}
		cOrsValue, err := fzf.FindOne(
			csName,
			//func(i int) string {
			//	return csName[i]
			//},
			func(item client.CSType) string {
				return string(item)
			},
			fzf.WithPrompt("Mount from Config/Secret >"),
		)
		if err != nil {
			return err
		}

		cOrs = client.CSType(*cOrsValue)
	}

	items := make([]server.ConfigORSecret, 0)
	if cOrs == client.ConfigType {
		configs, e := server.ListConfigs()

		if e != nil {
			return e
		}

		for _, c := range configs {
			items = append(items, server.ConfigORSecret{
				Entries: c.Data,
				Name:    c.Metadata.Name,
			})
		}

	} else {
		secrets, e := server.ListSecrets()

		if e != nil {
			return e
		}

		for _, c := range secrets {
			items = append(items, server.ConfigORSecret{
				Entries: c.StringData,
				Name:    c.Metadata.Name,
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
		selectedItemVal, err := fzf.FindOne(
			items,
			func(item server.ConfigORSecret) string {
				return item.Name
			},
			fzf.WithPrompt(fmt.Sprintf("Select %s >", cOrs)),
		)

		if err != nil {
			fn.PrintError(err)
		}

		selectedItem = *selectedItemVal
	}

	matchedIndex := -1
	for i, fe := range klFile.FileMount.Mounts {
		if fe.Path == path {
			matchedIndex = i
		}
	}

	key, err := fzf.FindOne(func() []string {
		res := make([]string, 0)
		for k := range selectedItem.Entries {
			res = append(res, k)
		}
		return res
	}(), func(item string) string {
		return item
	}, fzf.WithPrompt("Select Config/Secret >"))

	if err != nil {
		return err
	}

	if matchedIndex == -1 {
		klFile.FileMount.Mounts = append(klFile.FileMount.Mounts, client.FileEntry{
			Type: cOrs,
			Path: path,
			Name: selectedItem.Name,
			Key:  *key,
		})
	} else {
		klFile.FileMount.Mounts[matchedIndex] = client.FileEntry{
			Type: cOrs,
			Path: path,
			Name: selectedItem.Name,
			Key:  *key,
		}
	}

	if err := client.WriteKLFile(klFile); err != nil {
		return err
	}

	fmt.Println("Mount added to config")
	return nil
}

func init() {
	MountCommand.Flags().StringP("config", "", "", "config name")
	MountCommand.Flags().StringP("secret", "", "", "secret name")
}
