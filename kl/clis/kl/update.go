package kl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Hidden: true,
	Use:    "update",
	Short:  "Update the kl to latest version",
	Long: `Update the kl to latest version
Example:
# Update the kl to latest version
kl update
`,
	Run: func(cmd *cobra.Command, _ []string) {
		version := fn.ParseStringFlag(cmd, "version")

		if version != "" {
			version = fmt.Sprintf("@%s", version)
		}

		err := ExecUpdateCmd(version)
		if err != nil {
			fn.PrintError(err)
			return
		}
		fn.Log("successfully updated")
	},
}

func ExecUpdateCmd(version string) error {

	if runtime.GOOS == constants.RuntimeWindows {
		return fmt.Errorf("update is not supported on windows, please update manually using %q", text.Blue("iwr 'https://kl.kloudlite.io/kloudlite!?select=kl' | iex"))
	}

	var cmd *exec.Cmd
	curlAvailable := isCommandAvailable("curl")
	wgetAvailable := isCommandAvailable("wget")
	if curlAvailable {
		cmd = exec.Command("bash", "-c", fmt.Sprintf("curl %s%s!?select=%s | bash", constants.UpdateURL, version, flags.CliName))
	} else if wgetAvailable {
		cmd = exec.Command("bash", "-c", fmt.Sprintf("wget -qO - %s%s!?select=%s | bash", constants.UpdateURL, version, flags.CliName))
	} else {
		return fmt.Errorf("curl and wget not found")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func isCommandAvailable(command string) bool {
	cmd := exec.Command("which", command)
	err := cmd.Run()
	return err == nil
}

func init() {
	UpdateCmd.Flags().StringP("version", "v", "", fmt.Sprintf("%s cli version", flags.CliName))
}
