package shell

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var LocalShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "loading environment variables to current shell",
	Long: `This command will load default environment variables to the current shell
Example:
{cmd} shell
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		exec.Command("nix", "shell").Run()
		fmt.Println("shell")
	},
}

func init() {

}
