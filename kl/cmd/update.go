package cmd

import (
	"fmt"
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
	Run: func(_ *cobra.Command, _ []string) {
		err := ExecUpdateCmd()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("successfully updated")
	},
}

func ExecUpdateCmd() error {
	var cmd *exec.Cmd
	curlAvailable := isCommandAvailable("curl")
	wgetAvailable := isCommandAvailable("wget")
	if curlAvailable {
		cmd = exec.Command("bash", "-c", "curl https://i.jpillora.com/kloudlite/kl@v1.0.5-nightly! | bash")
	} else if wgetAvailable {
		cmd = exec.Command("bash", "-c", "wget -O- https://i.jpillora.com/kloudlite/kl@v1.0.5-nightly! | bash")
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
