package get

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/server"
	common_util "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "get config entries",
	Long: `get config entries
Examples:
  # get config entries in table
  kl get config <configid>

  # get config entries in json format
  kl get config <configid> -o json

  # get config entries in yaml format
  kl get config <configid> -o yaml
`,
	Run: func(cmd *cobra.Command, args []string) {
		configId := ""
		if len(args) >= 1 {
			configId = args[0]
		} else {
			var err error
			configId, err = selectConfig()
			if err != nil {
				common_util.PrintError(err)
				return
			}
		}

		err := printConfig(configId, cmd)
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}

func selectConfig() (string, error) {
	var err error
	configs, err := server.GetConfigs()

	if len(configs) == 0 {
		return "", errors.New("no configs present in your current project")
	}

	selectedIndex, err := fuzzyfinder.Find(
		configs,
		func(i int) string {
			return configs[i].Name
		},
		fuzzyfinder.WithPromptString("Select Config >"),
	)

	if err != nil {
		return "", err
	}

	return configs[selectedIndex].Id, nil
}

func printConfig(configId string, cmd *cobra.Command) error {
	outputFormat := cmd.Flag("output").Value.String()

	config, err := server.GetConfig(configId)
	if err != nil {
		return err
	}

	switch outputFormat {
	case "json":
		configBytes, err := json.Marshal(config)
		if err != nil {
			return err
		}
		fmt.Println(string(configBytes))

	case "yaml", "yml":
		configBytes, err := yaml.Marshal(config)
		if err != nil {
			return err
		}
		fmt.Println(string(configBytes))

	default:
		header := table.Row{
			table.HeaderText("key"),
			table.HeaderText("value"),
		}
		rows := make([]table.Row, 0)

		for _, c := range config.Entries {
			rows = append(rows, table.Row{
				c.Key, c.Value,
			})
		}

		fmt.Println(table.Table(&header, rows))

		fmt.Println(
			table.KVOutput("Showing entries of config:", config.Name, true),
		)

		table.TotalResults(len(config.Entries), true)

	}

	return nil
}

func init() {
	configCmd.Flags().StringP("output", "o", "table", "json | yaml")
}
