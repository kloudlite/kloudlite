package list

import (
	"fmt"
	"github.com/kloudlite/kl/domain/fileclient"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Get list of secrets in selected environment",
	Run: func(cmd *cobra.Command, args []string) {
		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		filePath := fn.ParseKlFile(cmd)
		klFile, err := fc.GetKlFile(filePath)
		if err != nil {
			fn.PrintError(err)
			return
		}

		sec, err := apiclient.ListSecrets([]fn.Option{
			fn.MakeOption("accountName", klFile.AccountName),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printSecrets(cmd, sec); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printSecrets(_ *cobra.Command, secrets []apiclient.Secret) error {
	if len(secrets) == 0 {
		return functions.Error("no secrets found")
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		table.HeaderText("entries"),
	}

	rows := make([]table.Row, 0)

	for _, a := range secrets {
		rows = append(rows, table.Row{a.DisplayName, a.Metadata.Name, fmt.Sprintf("%d", len(a.StringData))})
	}

	fmt.Println(table.Table(&header, rows))
	table.TotalResults(len(secrets), true)
	return nil
}

func init() {
	secretsCmd.Aliases = append(secretsCmd.Aliases, "secret")
	secretsCmd.Aliases = append(secretsCmd.Aliases, "sec")
}
