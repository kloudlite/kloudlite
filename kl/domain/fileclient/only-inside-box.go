package fileclient

import (
	"fmt"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

func OnlyInsideBox(cmd *cobra.Command) {
	if InsideBox() {
		return
	}

	cmd.Run = func(*cobra.Command, []string) {
		fn.PrintError(fmt.Errorf(`must be executed inside a development container.`))
		return
	}
}

func OnlyOutsideBox(cmd *cobra.Command) {
	if !InsideBox() {
		return
	}

	cmd.Run = func(*cobra.Command, []string) {
		fn.PrintError(fmt.Errorf("must be executed on host machine."))
		return
	}
}
