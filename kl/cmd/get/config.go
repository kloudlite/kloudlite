package get

import (
	"encoding/json"
	"fmt"

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

		configName := ""

		if len(args) >= 1 {
			configName = args[0]
		}
		// filePath := fn.ParseKlFile(cmd)
		// klFile, err := fc.GetKlFile(filePath)
		// if err != nil {
		// 	fn.PrintError(err)
		// 	return
		// }

		// config, err := apic.EnsureConfig([]fn.Option{
		// 	fn.MakeOption("configName", configName),
		// 	fn.MakeOption("accountName", klFile.AccountName),
		// }...)
		// if err != nil {
		// 	fn.PrintError(err)
		// 	return
		// }

		config, err := apic.GetConfig(fn.MakeOption("configName", configName))
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

		fmt.Println(table.Table(&header, rows))

		table.KVOutput("Showing entries of config:", config.Metadata.Name, true)

		table.TotalResults(len(config.Data), true)
	}

	return nil
}

func init() {
	configCmd.Flags().StringP("output", "o", "table", "json | yaml")
}
