package list

import (
	"errors"
	"fmt"

	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/kloudlite/kl/lib/ui/table"
	"github.com/kloudlite/kl/lib/util"
	"github.com/spf13/cobra"
)

var configsCmd = &cobra.Command{
	Use:   "configs",
	Short: "list all the configs available in selected project",
	Long: `List all the configs available in project.
Examples:
	# list all the configs with selected project
  kl list configs

	# list all the configs with projectId
  kl list configs <projectId>
`,
	Run: func(_ *cobra.Command, args []string) {
		err := listConfigs(args)
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}

func listConfigs(args []string) error {

	var configs []server.Config
	var err error
	projectId := ""

	if len(args) >= 1 {
		projectId = args[0]
	}

	if projectId == "" {
		configs, err = server.GetConfigs()
	} else {
		configs, err = server.GetConfigs(common_util.MakeOption("projectId", args[0]))
	}

	if err != nil {
		return err
	}

	if len(configs) == 0 {
		return errors.New("no configs found")
	}

	header := table.Row{
		table.HeaderText("configs"),
		table.HeaderText("id"),
		table.HeaderText("entries"),
	}

	rows := make([]table.Row, 0)

	for _, a := range configs {
		rows = append(rows,
			table.Row{
				a.Name,
				a.Id,
				fmt.Sprintf("%d entries",
					len(a.Entries)),
			},
		)
	}

	fmt.Println(table.Table(&header, rows))

	if projectId == "" {
		projectId, _ = util.CurrentProjectName()
	}

	if projectId != "" {
		table.KVOutput("configs of", projectId, true)
	}

	table.TotalResults(len(configs), true)

	return nil
}

func init() {
	configsCmd.Aliases = append(configsCmd.Aliases, "config")
	configsCmd.Aliases = append(configsCmd.Aliases, "conf")
}
