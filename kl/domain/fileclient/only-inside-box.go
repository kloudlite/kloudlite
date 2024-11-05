package fileclient

import (
	"github.com/kloudlite/kl/domain/envclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

func OnlyInsideBox(cmd *cobra.Command) {
	if envclient.InsideBox() {
		return
	}

	cmd.Run = func(*cobra.Command, []string) {
		fn.PrintError(fn.Errorf(`must be executed inside a development container.`))
		return
	}
}

func OnlyOutsideBox(cmd *cobra.Command) {
	if !envclient.InsideBox() {
		return
	}

	cmd.Run = func(*cobra.Command, []string) {
		fn.PrintError(fn.Errorf("must be executed on host machine."))
		return
	}
}
