package box

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/spf13/cobra"
)

var BoxCmd = &cobra.Command{
	Use:   "box",
	Short: "start, stop, reload, ssh and get running box info",
}

func init() {

	// fileclient.OnlyInsideBox(reloadCmd)
	BoxCmd.AddCommand(reloadCmd)

	fileclient.OnlyOutsideBox(startCmd)
	BoxCmd.AddCommand(startCmd)

	fileclient.OnlyOutsideBox(sshCmd)
	BoxCmd.AddCommand(sshCmd)

	fileclient.OnlyOutsideBox(execCmd)
	BoxCmd.AddCommand(execCmd)

	fileclient.OnlyOutsideBox(stopCmd)
	BoxCmd.AddCommand(stopCmd)

	fileclient.OnlyOutsideBox(psCmd)
	BoxCmd.AddCommand(psCmd)

	fileclient.OnlyOutsideBox(restartCmd)
	BoxCmd.AddCommand(restartCmd)

	BoxCmd.AddCommand(infoCmd)
}
