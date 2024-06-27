package shell

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/constants"

	domain_util "github.com/kloudlite/kl/domain/util"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"

	// "path"
	"strings"
)

var ShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "loading environment variables to current shell",
	Long: `This command will load default environment variables to the current shell
Example:
{cmd} shell
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		if err := loadEnv(cmd); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func loadEnv(cmd *cobra.Command) error {
	var err error

	shell, err := ShellName()
	if err != nil {
		return functions.NewE(err)
	}

	if err := domain_util.MountEnv([]string{"--", shell}); err != nil {
		return functions.NewE(err)
	}

	return nil
}

func ShellName() (string, error) {

	shell, exists := os.LookupEnv("SHELL")
	if !exists {
		return "", fmt.Errorf("shell not found")
	}
	if strings.Contains(shell, constants.BashShell) {
		return constants.BashShell, nil
	} else if strings.Contains(shell, constants.FishShell) {
		return constants.FishShell, nil
	} else if strings.Contains(shell, constants.ZshShell) {
		return constants.ZshShell, nil
	} else if strings.Contains(shell, constants.PowerShell) {
		return constants.PowerShell, nil
	}
	return "", fmt.Errorf("unsupported shell")
}

func init() {
	ShellCmd.Aliases = append(ShellCmd.Aliases, "s", "sh")
	ShellCmd.Flags().StringP("account", "a", "", "account name")
	ShellCmd.Flags().StringP("cluster", "c", "", "cluster name")
}
