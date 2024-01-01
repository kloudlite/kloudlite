package use

import (
	"fmt"
	common_util "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

	common_cmd "github.com/kloudlite/kl/cmd/common"
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

  # select account with cluster id
  kl use cluster <clusterId>
	`,
	Run: func(_ *cobra.Command, args []string) {
		clusterName, err := common_cmd.SelectCluster(args)
		if err != nil {
			common_util.PrintError(err)
			return
		}

		fmt.Println(text.Bold(text.Green("\nSelected cluster:")),
			text.Blue(fmt.Sprintf("%s (%s)", clusterName.DisplayName, clusterName.Name)),
		)
	},
}
