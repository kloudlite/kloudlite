package cluster

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cluster",
	Short: "list of clusters",
	Long: `Using this command you can list all the clusters.
Examples:
  # list clusters accessible to you
  kl infra clusters list
`,
}

func init() {

	Cmd.Aliases = append(Cmd.Aliases, "cluster")
	Cmd.Aliases = append(Cmd.Aliases, "clus")
	Cmd.AddCommand(clustersCmd)

}
