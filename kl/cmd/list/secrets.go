package list

import (
	"errors"
	"fmt"

	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/kloudlite/kl/lib/ui/table"
	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "list all the secrets available in selected project",
	Long: `List all the secrets available in project.
Examples:
	# list all the secrets with selected project
  kl list secrets

	# list all the secrets with projectId
  kl list secrets <projectId>
`,
	Run: func(_ *cobra.Command, args []string) {
		err := listSecrets(args)
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}

func listSecrets(args []string) error {

	var secrets []server.Secret
	var err error
	projectId := ""

	if len(args) >= 1 {
		projectId = args[0]
	}

	if projectId == "" {
		secrets, err = server.GetSecrets()
	} else {
		secrets, err = server.GetSecrets(common_util.MakeOption("projectId", args[0]))
	}

	if err != nil {
		return err
	}

	if len(secrets) == 0 {
		return errors.New("no secrets found")
	}

	header := table.Row{
		table.HeaderText("secrets"),
		table.HeaderText("id"),
		table.HeaderText("entries"),
	}

	rows := make([]table.Row, 0)

	for _, a := range secrets {
		rows = append(rows, table.Row{a.Name, a.Id, fmt.Sprintf("%d", len(a.Entries))})
	}

	fmt.Println(table.Table(&header, rows))

	if projectId == "" {
		projectId, _ = server.CurrentProjectId()
	}

	if projectId != "" {
		table.KVOutput("secrets of", projectId, true)
	}

	table.TotalResults(len(secrets), true)

	return nil
}

func init() {
	secretsCmd.Aliases = append(secretsCmd.Aliases, "secret")
	secretsCmd.Aliases = append(secretsCmd.Aliases, "sec")
}
