package get

import (
	"encoding/json"
	"fmt"
	"github.com/kloudlite/kl/domain/fileclient"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var secretCmd = &cobra.Command{
	Use:   "secret [name]",
	Short: "list secrets entries",
	Long:  "use this command to list the entries of specific secret",
	Run: func(cmd *cobra.Command, args []string) {
		secName := ""

		if len(args) >= 1 {
			secName = args[0]
		}

		filePath := fn.ParseKlFile(cmd)
		klFile, err := fileclient.GetKlFile(filePath)
		if err != nil {
			fn.PrintError(err)
			return
		}

		sec, err := apiclient.EnsureSecret([]fn.Option{
			fn.MakeOption("secretName", secName),
			fn.MakeOption("accountName", klFile.AccountName),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printSecret(sec, cmd); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printSecret(secret *apiclient.Secret, cmd *cobra.Command) error {
	outputFormat := cmd.Flag("output").Value.String()

	switch outputFormat {
	case "json":
		configBytes, err := json.Marshal(secret.StringData)
		if err != nil {
			return functions.NewE(err)
		}
		fn.Println(string(configBytes))

	case "yaml", "yml":
		configBytes, err := yaml.Marshal(secret.StringData)
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

		for k, v := range secret.StringData {
			rows = append(rows, table.Row{
				k, v,
			})
		}

		fmt.Println(table.Table(&header, rows))
		table.KVOutput("Showing entries of secret:", secret.Metadata.Name, true)
		table.TotalResults(len(secret.StringData), true)
	}

	return nil
}

func init() {
	secretCmd.Flags().StringP("output", "o", "table", "output format (table|json|yaml)")
	secretCmd.Aliases = append(secretCmd.Aliases, "sec")
}
