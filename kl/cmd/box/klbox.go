package box

import (
	"github.com/spf13/cobra"
)

var BoxCmd = &cobra.Command{
	Use:   "box",
	Short: "box [start | stop | ssh | restart]",
	Long: ` Use this command to start, stop, restart and open an ssh connection to the container with loaded envs
kl box start
kl box stop
kl box ssh
kl box restart
`,
}

func init() {
	BoxCmd.AddCommand(startCmd)
	BoxCmd.AddCommand(stopCmd)
	BoxCmd.AddCommand(sshCmd)
	BoxCmd.AddCommand(restartCmd)
	BoxCmd.AddCommand(execCmd)
	//BoxCmd.Aliases = append(BoxCmd.Aliases, "b")
}
