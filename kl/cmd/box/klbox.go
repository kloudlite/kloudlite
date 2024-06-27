package box

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/spf13/cobra"
)

var BoxCmd = &cobra.Command{
	Use:   "box",
	Short: "start, stop, reload, ssh and get running box info",
	//PersistentPreRun: func(cmd *cobra.Command, args []string) {
	//	apiclient.EnsureBoxHash()
	//},
}

func init() {

	//fileclient.OnlyInsideBox(reloadCmd)
	BoxCmd.AddCommand(reloadCmd)

	fileclient.OnlyOutsideBox(startCmd)
	BoxCmd.AddCommand(startCmd)

	fileclient.OnlyOutsideBox(stopAllCmd)
	BoxCmd.AddCommand(stopAllCmd)

	fileclient.OnlyOutsideBox(sshCmd)
	BoxCmd.AddCommand(sshCmd)

	fileclient.OnlyOutsideBox(execCmd)
	BoxCmd.AddCommand(execCmd)

	fileclient.OnlyOutsideBox(stopCmd)
	BoxCmd.AddCommand(stopCmd)

	fileclient.OnlyOutsideBox(psCmd)
	BoxCmd.AddCommand(psCmd)

	BoxCmd.AddCommand(infoCmd)
}

func setBoxCommonFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("foreground", "f", false, "run in foreground mode")
}
