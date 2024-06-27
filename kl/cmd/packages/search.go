package packages

import (
	"fmt"
	"strings"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [name]",
	Short: "search for a package",
	Run: func(cmd *cobra.Command, args []string) {
		if err := searchPackages(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func searchPackages(cmd *cobra.Command, args []string) error {
	name := fn.ParseStringFlag(cmd, "name")
	if name == "" && len(args) > 0 {
		name = args[0]
	}

	if name == "" {
		return functions.Error("name is required")
	}

	stopSp := spinner.Client.Start(fmt.Sprintf("searching for package %s", name))
	defer stopSp()

	sr, err := Search(cmd.Context(), name)
	if err != nil {
		return functions.NewE(err)
	}
	stopSp()

	header := table.Row{table.HeaderText("name"), table.HeaderText("versions")}
	rows := make([]table.Row, 0)

	for _, p := range sr.Packages {
		versions := make([]string, 0)
		for j, v := range p.Versions {
			if j >= 10 {
				break
			}

			versions = append(versions, v.Version)
		}

		rows = append(rows, table.Row{
			text.Bold(text.Green(p.Name)),
			fmt.Sprintf("%s", strings.Join(versions, ", ")),
		})
	}

	fmt.Println(table.Table(&header, rows, cmd))

	return nil
}

func init() {
	searchCmd.Flags().StringP("name", "n", "", "name of the package to remove")
	searchCmd.Flags().BoolP("show-all", "a", false, "list all matching packages")
}
