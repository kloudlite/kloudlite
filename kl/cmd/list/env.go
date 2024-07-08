package list

import (
	"fmt"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "envs",
	Short: "Get list of environments",
	Run: func(cmd *cobra.Command, args []string) {
		err := listEnvironments(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listEnvironments(cmd *cobra.Command, args []string) error {

	fc, err := fileclient.New()
	if err != nil {
		return functions.NewE(err)
	}

	apic, err := apiclient.New()
	if err != nil {
		return functions.NewE(err)
	}

	klFile, err := fc.GetKlFile("")
	if err != nil {
		return functions.NewE(err)
	}
	envs, err := apic.ListEnvs(fn.MakeOption("accountName", klFile.AccountName))
	if err != nil {
		return functions.NewE(err)
	}

	if len(envs) == 0 {
		return functions.Error("no environments found")
	}

	env, _ := fc.CurrentEnv()
	envName := ""
	if env != nil {
		envName = env.Name
	}

	header := table.Row{table.HeaderText("Display Name"), table.HeaderText("Name"), table.HeaderText("ready")}
	rows := make([]table.Row, 0)

	for _, a := range envs {
		rows = append(rows, table.Row{
			fn.GetPrintRow(a, envName, a.DisplayName, true),
			fn.GetPrintRow(a, envName, a.Metadata.Name),
			fn.GetPrintRow(a, envName, a.Status.IsReady),
		})
	}

	fmt.Println(table.Table(&header, rows, cmd))

	if s := fn.ParseStringFlag(cmd, "output"); s == "table" {
		table.TotalResults(len(envs), true)
	}

	return nil
}

func init() {
	envCmd.Aliases = append(envCmd.Aliases, "env")
}
