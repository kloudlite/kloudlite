package box

import (
	"github.com/spf13/cobra"
)

var BoxCmd = &cobra.Command{
	Use:   "box",
	Short: "box [start | stop | ssh | restart | exec]",
}

func init() {
	BoxCmd.AddCommand(startCmd)
	BoxCmd.AddCommand(stopCmd)
	BoxCmd.AddCommand(stopAllCmd)
	BoxCmd.AddCommand(sshCmd)
	BoxCmd.AddCommand(execCmd)
	BoxCmd.AddCommand(restartCmd)
	BoxCmd.AddCommand(psCmd)
}

func setBoxCommonFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("verbose", "v", false, "run in verbose mode")
	cmd.Flags().BoolP("foreground", "f", false, "run in foreground mode")
}
