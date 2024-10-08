package cluster

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage clusters",
	Long:  `start and stop clusters.`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "clus", "clusters")

	UpCmd.Aliases = append(UpCmd.Aliases, "start", "connect", "create")
	DownCmd.Aliases = append(DownCmd.Aliases, "stop", "disconnect", "destroy")

	fileclient.OnlyOutsideBox(UpCmd)
	fileclient.OnlyOutsideBox(DownCmd)
	Cmd.AddCommand(DownCmd)
	Cmd.AddCommand(UpCmd)
}
