//go:build main

package apploader

import (
	"github.com/kloudlite/kl/cmd/app"
	"github.com/spf13/cobra"
)

func LoadStartApp(root *cobra.Command) {
	// if runtime.GOOS == constants.RuntimeWindows {
	// 	return
	// }

	root.AddCommand(app.Cmd)
}
