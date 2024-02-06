package kli

import (
	"fmt"
	"github.com/kloudlite/kl/constants"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
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
		version := ""

		if cmd.Flags().Changed("version") {
			version, _ = cmd.Flags().GetString("version")
			version = fmt.Sprintf("@%s", version)
		}
		err := ExecUpdateCmd(version)
		if err != nil {
			fn.Log(err)
			return
		}
		fn.Log("successfully updated")
	},
}

func ExecUpdateCmd(version string) error {
	var cmd *exec.Cmd
	curlAvailable := isCommandAvailable("curl")
	wgetAvailable := isCommandAvailable("wget")
	if curlAvailable {
		cmd = exec.Command("bash", "-c", fmt.Sprintf("curl '%s%s!?source=%s' | bash", constants.UpdateURL, version, constants.InfraCliName))
	} else if wgetAvailable {
		cmd = exec.Command("bash", "-c", fmt.Sprintf("wget -qO - '%s%s!?source=%s' | bash", constants.UpdateURL, version, constants.InfraCliName))
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
	UpdateCmd.Flags().StringP("version", "v", "", "kl cli version")
}
