package kl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/updater"
	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the kl to latest version",
	Long: `Update the kl to latest version
Example:
# Update the kl to latest version
kl update
`,
	Run: func(cmd *cobra.Command, args []string) {
		version := ""
		if len(args) > 0 {
			version = args[0]
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
	uurl, err := updater.NewUpdater().GetUpdateUrl()
	if err != nil {
		uurl = &constants.UpdateURL
	}

	if runtime.GOOS == constants.RuntimeWindows {
		return fn.Errorf("update is not supported on windows, please update manually using %q", text.Blue("iwr 'https://kl.kloudlite.io/kloudlite!?select=kl' | iex"))
	}

	var cmd *exec.Cmd
	curlAvailable := isCommandAvailable("curl")
	wgetAvailable := isCommandAvailable("wget")
	if curlAvailable {
		cmd = exec.Command("bash", "-c", fmt.Sprintf("curl %s%s!?select=%s | bash", *uurl, version, flags.CliName))
	} else if wgetAvailable {
		cmd = exec.Command("bash", "-c", fmt.Sprintf("wget -qO - %s%s!?select=%s | bash", *uurl, version, flags.CliName))
	} else {
		return fn.Errorf("curl and wget not found")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		return functions.NewE(err)
	}
	return nil
}

func isCommandAvailable(command string) bool {
	cmd := exec.Command("which", command)
	err := cmd.Run()
	return functions.NewE(err) == nil
}
