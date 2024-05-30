package packages

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/domain/client"
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

	verbose := fn.ParseBoolFlag(cmd, "verbose")

	stopSp := spinner.Client.Start(fmt.Sprintf("removing package %s", name))
	defer stopSp()

	err := client.ExecPackageCommand(fmt.Sprintf("devbox rm %s%s", name, func() string {
		if verbose {
			return ""
		}
		return " -q"
	}()))
	stopSp()
	if err != nil {
		return err
	}

	fn.Logf("removed package %s", name)
	return nil
}

func init() {
	rmCmd.Flags().StringP("name", "n", "", "name of the package to remove")
	rmCmd.Flags().BoolP("verbose", "v", false, "name of the package to install")
}
