package packages

import (
	"errors"
	"fmt"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add new package",
	Run: func(cmd *cobra.Command, args []string) {
		if err := addPackages(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func addPackages(cmd *cobra.Command, args []string) error {
	name := fn.ParseStringFlag(cmd, "name")
	if name == "" && len(args) > 0 {
		name = args[0]
	}

	if name == "" {
		return errors.New("name is required")
	}

	stopSp := spinner.Client.Start(fmt.Sprintf("adding package %s", name))
	defer stopSp()

	err := execPackageCommand(fmt.Sprintf("devbox add %s -q", name))
	stopSp()
	if err != nil {
		return err
	}

	fn.Logf("added package %s", name)
	return nil
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "name of the package to install")
}
