package add

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/domain/apiclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"

	"github.com/spf13/cobra"
)

var mountCommand = &cobra.Command{
	Use:   "config-mount [path]",
	Short: "add file mount to your kl-config file by selection from the all the [ config | secret ] available in current environemnt",
	Long: `
	This command will help you to add mounts to your kl-config file.
	You can add a config or secret to your kl-config file by providing the path of the config/secret you want to add.
	`,
	Example: `
  kl add config-mount [path] --config=<config_name>	# add mount from config.
  kl add config-mount [path] --secret=<secret_name>	# add secret from secret.
`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath := fn.ParseKlFile(cmd)

		klFile, err := fileclient.GetKlFile(filePath)
		if err != nil {
			fn.PrintError(err)
			return
		}
		path := ""

		if len(args) > 0 {
			path = args[0]
		} else {
			fn.PrintError(fmt.Errorf("please specify the path of the config you want to add, example: kl add config-mount /tmp/sample"))
			return
		}

		err = selectConfigMount(path, *klFile, cmd)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func selectConfigMount(path string, klFile fileclient.KLFileType, cmd *cobra.Command) error {
	//TODO: add changes to the klbox-hash file
	c := cmd.Flag("config").Value.String()
	s := cmd.Flag("secret").Value.String()

	var cOrs fileclient.CSType
	cOrs = ""

	if c != "" || s != "" {

		if c != "" {
			cOrs = fileclient.ConfigType
		} else {
			cOrs = fileclient.SecretType
		}

	} else {
		csName := []fileclient.CSType{fileclient.ConfigType, fileclient.SecretType}
		cOrsValue, err := fzf.FindOne(
			csName,
			//func(i int) string {
			//	return csName[i]
			//},
			func(item fileclient.CSType) string {
				return string(item)
			},
			fzf.WithPrompt("Mount from Config/Secret >"),
		)
		if err != nil {
			return fn.NewE(err)
		}

		cOrs = fileclient.CSType(*cOrsValue)
	}

	items := make([]apiclient.ConfigORSecret, 0)
	if cOrs == fileclient.ConfigType {
		configs, e := apiclient.ListConfigs([]fn.Option{
			fn.MakeOption("accountName", klFile.AccountName),
		}...)

		if e != nil {
			return e
		}

		for _, c := range configs {
			items = append(items, apiclient.ConfigORSecret{
				Entries: c.Data,
				Name:    c.Metadata.Name,
			})
		}

	} else {
		secrets, e := apiclient.ListSecrets([]fn.Option{
			fn.MakeOption("accountName", klFile.AccountName),
		}...)

		if e != nil {
			return e
		}

		for _, c := range secrets {
			items = append(items, apiclient.ConfigORSecret{
				Entries: c.StringData,
				Name:    c.Metadata.Name,
			})
		}
	}

	if len(items) == 0 {
		return fmt.Errorf("no %ss created yet on apiclient ", cOrs)
	}

	selectedItem := apiclient.ConfigORSecret{}

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
			func(item apiclient.ConfigORSecret) string {
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
	for i, fe := range klFile.Mounts {
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
		return fn.NewE(err)
	}

	fe := klFile.Mounts.GetMounts()

	if matchedIndex == -1 {
		fe = append(fe, fileclient.FileEntry{
			Type: cOrs,
			Path: path,
			Name: selectedItem.Name,
			Key:  *key,
		})
	} else {
		fe[matchedIndex] = fileclient.FileEntry{
			Type: cOrs,
			Path: path,
			Name: selectedItem.Name,
			Key:  *key,
		}
	}

	klFile.Mounts.AddMounts(fe)
	if err := fileclient.WriteKLFile(klFile); err != nil {
		return fn.NewE(err)
	}

	fn.Log("added mount to your kl-file")

	wpath, err := os.Getwd()
	if err != nil {
		return fn.NewE(err)
	}

	if err = hashctrl.SyncBoxHash(wpath); err != nil {
		return fn.NewE(err)
	}

	return nil
}

func init() {
	mountCommand.Flags().StringP("config", "", "", "config name")
	mountCommand.Flags().StringP("secret", "", "", "secret name")
	fn.WithKlFile(mountCommand)
}
