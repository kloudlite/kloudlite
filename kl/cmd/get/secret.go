package get

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/table"
	"github.com/kloudlite/kl/lib/server"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "get secret entries",
	Long: `get secret entries
Examples:
  # get secret entries in table
  kl get secret <secretid>

  # get secret entries in json format
  kl get secret <secretid> -o json

  # get secret entries in yaml format
  kl get secret <secretid> -o yaml
`,
	Run: func(cmd *cobra.Command, args []string) {
		secretId := ""
		if len(args) >= 1 {
			secretId = args[0]
		} else {
			var err error
			secretId, err = selectSecret()
			if err != nil {
				common.PrintError(err)
				return
			}
		}

		err := printSecret(secretId, cmd)
		if err != nil {
			common.PrintError(err)
			return
		}

	},
}

func selectSecret() (string, error) {
	var err error
	secrets, err := server.GetSecrets()

	if len(secrets) == 0 {
		return "", errors.New("no secret present in your current project")
	}

	selectedIndex, err := fuzzyfinder.Find(
		secrets,
		func(i int) string {
			return secrets[i].Name
		},
		fuzzyfinder.WithPromptString("Select Secret >"),
	)

	if err != nil {
		return "", err
	}

	return secrets[selectedIndex].Name, nil
}

func printSecret(secretId string, cmd *cobra.Command) error {
	outputFormat := cmd.Flag("output").Value.String()

	secret, err := server.GetSecret(secretId)
	if err != nil {
		return err
	}

	switch outputFormat {
	case "json":
		secretBytes, err := json.Marshal(secret)
		if err != nil {
			return err
		}
		fmt.Println(string(secretBytes))

	case "yaml", "yml":

		secretBytes, err := yaml.Marshal(secret)
		if err != nil {
			return err
		}
		fmt.Println(string(secretBytes))

	default:
		header := table.Row{
			table.HeaderText("key"),
			table.HeaderText("value"),
		}
		rows := make([]table.Row, 0)

		for _, c := range secret.Entries {
			rows = append(rows, table.Row{
				c.Key, c.Value,
			})
		}

		fmt.Println(table.Table(&header, rows))

		table.KVOutput("Showing entries of secret:", secret.Name, true)

		table.TotalResults(len(secret.Entries), true)

	}

	return nil
}

func init() {
	secretCmd.Flags().StringP("output", "o", "table", "output format (table|json|yaml)")
}
