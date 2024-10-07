package get

import (
	"encoding/json"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/ui/fzf"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var configCmd = &cobra.Command{
	Use:   "config [name]",
	Short: "list config entries",
	Long:  "use this command to list entries of specific config",
	Run: func(cmd *cobra.Command, args []string) {

		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		configName := ""

		if len(args) >= 1 {
			configName = args[0]
		}

		if configName == "" {
			currentTeam, err := fc.CurrentTeamName()
			if err != nil {
				fn.PrintError(err)
				return
			}
			currentEnv, err := apic.EnsureEnv()
			if err != nil {
				fn.PrintError(err)
				return
			}
			configs, err := apic.ListConfigs(currentTeam, currentEnv.Name)
			if err != nil {
				fn.PrintError(err)
				return
			}
			if len(configs) == 0 {
				fn.PrintError(fn.Error("no configs found"))
				return
			}
			selectedConfig, err := fzf.FindOne(configs, func(config apiclient.Config) string {
				return config.DisplayName
			}, fzf.WithPrompt("select config > "))
			if err != nil {
				fn.PrintError(err)
				return
			}
			configName = selectedConfig.Metadata.Name
		}

		currentTeamName, err := fc.CurrentTeamName()
		if err != nil {
			fn.PrintError(err)
			return
		}
		currentEnvName, err := apic.EnsureEnv()
		if err != nil {
			fn.PrintError(err)
			return
		}

		config, err := apic.GetConfig(currentTeamName, currentEnvName.Name, configName)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printConfig(config, cmd); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printConfig(config *apiclient.Config, cmd *cobra.Command) error {
	outputFormat := cmd.Flag("output").Value.String()

	switch outputFormat {
	case "json":
		configBytes, err := json.Marshal(config.Data)
		if err != nil {
			return functions.NewE(err)
		}
		fn.Println(string(configBytes))

	case "yaml", "yml":
		configBytes, err := yaml.Marshal(config.Data)
		if err != nil {
			return functions.NewE(err)
		}
		fn.Println(string(configBytes))

	default:
		header := table.Row{
			table.HeaderText("key"),
			table.HeaderText("value"),
		}
		rows := make([]table.Row, 0)

		for k, v := range config.Data {
			rows = append(rows, table.Row{
				k, v,
			})
		}

		fn.Println(table.Table(&header, rows))

		table.KVOutput("Showing entries of config:", config.Metadata.Name, true)

		table.TotalResults(len(config.Data), true)
	}

	return nil
}

func init() {
	configCmd.Flags().StringP("output", "o", "table", "json | yaml")
}
