package use

import (
	"fmt"

	"github.com/kloudlite/kl/domain/server"
	common_util "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "select cluster to use later with all commands",
	Long: `Select cluster
Examples:
  # select cluster
  kl use cluster
	# this will open selector where you can select one of the cluster accessible to you.

  # select account with cluster name
  kl use cluster <clustername>
	`,
	Run: func(_ *cobra.Command, args []string) {

		clusterName := ""
		if len(args) >= 1 {
			clusterName = args[0]
		}

		cluster, err := server.SelectCluster(clusterName)
		if err != nil {
			common_util.PrintError(err)
			return
		}

		fmt.Println(text.Bold(text.Green("\nSelected cluster:")),
			text.Blue(fmt.Sprintf("%s (%s)", cluster.DisplayName, cluster.Metadata.Name)),
		)
	},
}
