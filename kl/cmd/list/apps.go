package list

import (
	"strconv"
	"strings"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Get list of apps in selected environment",
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

		if err := listapps(apic, fc, cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listapps(apic apiclient.ApiClient, fc fileclient.FileClient, cmd *cobra.Command, _ []string) error {
	fc, err := fileclient.New()
	if err != nil {
		return functions.NewE(err)
	}

	envName := fn.ParseStringFlag(cmd, "env")

	currentAccountName, err := fc.CurrentAccountName()
	if err != nil {
		return functions.NewE(err)
	}
	currentEnvName, err := fc.CurrentEnv()
	if err != nil {
		return functions.NewE(err)
	}

	apps, err := apic.ListApps(currentAccountName, currentEnvName.Name)
	if err != nil {
		return functions.NewE(err)
	}

	if len(apps) == 0 {
		return functions.Error("no apps found")
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		table.HeaderText("App Port"),
	}

	rows := make([]table.Row, 0)

	ports := make([]string, 0)
	for _, a := range apps {
		ports = nil
		for _, v := range a.Spec.Services {
			ports = append(ports, strconv.Itoa(v.Port))
		}
		rows = append(rows, table.Row{a.DisplayName, a.Metadata.Name, strings.Join(ports, ", ")})
	}

	fn.Println(table.Table(&header, rows, cmd))

	table.KVOutput("apps of", envName, true)
	table.TotalResults(len(apps), true)
	return nil
}

func init() {
	appsCmd.Aliases = append(appsCmd.Aliases, "app")
	appsCmd.Flags().StringP("env", "e", "", "environment name")
}
