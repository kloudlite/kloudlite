package packages

import (
	"errors"
	"fmt"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "remove installed package",
	Run: func(cmd *cobra.Command, args []string) {
		if err := rmPackages(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func rmPackages(cmd *cobra.Command, args []string) error {
	name := fn.ParseStringFlag(cmd, "name")
	if name == "" && len(args) > 0 {
		name = args[0]
	}

	if name == "" {
		return errors.New("name is required")
	}

	stopSp := spinner.Client.Start(fmt.Sprintf("removing package %s", name))
	defer stopSp()

	err := execPackageCommand(fmt.Sprintf("devbox rm %s -q", name))
	stopSp()
	if err != nil {
		return err
	}

	fn.Logf("removed package %s", name)
	return nil
}

func init() {
	rmCmd.Flags().StringP("name", "n", "", "name of the package to remove")
}
