package packages

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/domain/client"
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

	verbose := fn.ParseBoolFlag(cmd, "verbose")

	stopSp := spinner.Client.Start(fmt.Sprintf("adding package %s", name))
	defer stopSp()

	err := client.ExecPackageCommand(fmt.Sprintf("devbox add %s%s", name, func() string {
		if verbose {
			return ""
		}
		return " -q"
	}()))
	stopSp()
	if err != nil {
		return err
	}

	fn.Logf("added package %s", name)

	return nil
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "name of the package to install")
	addCmd.Flags().BoolP("verbose", "v", false, "name of the package to install")
}
