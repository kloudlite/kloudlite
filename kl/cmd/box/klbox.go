package box

import (
	"github.com/kloudlite/kl/domain/client"
	"github.com/spf13/cobra"
)

var BoxCmd = &cobra.Command{
	Use:   "box",
	Short: "box [start | stop | ssh | reload | ps]",
}

func init() {

	client.OnlyInsideBox(reloadCmd)
	BoxCmd.AddCommand(reloadCmd)

	client.OnlyOutsideBox(startCmd)
	BoxCmd.AddCommand(startCmd)

	client.OnlyOutsideBox(stopAllCmd)
	BoxCmd.AddCommand(stopAllCmd)

	client.OnlyOutsideBox(sshCmd)
	BoxCmd.AddCommand(sshCmd)

	client.OnlyOutsideBox(execCmd)
	BoxCmd.AddCommand(execCmd)

	client.OnlyOutsideBox(psCmd)
	BoxCmd.AddCommand(stopCmd)

	client.OnlyOutsideBox(psCmd)
	BoxCmd.AddCommand(psCmd)
}

func setBoxCommonFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("verbose", "v", false, "run in verbose mode")
	cmd.Flags().BoolP("foreground", "f", false, "run in foreground mode")
}
