//go:build main

package apploader

import (
	app "github.com/kloudlite/kl/cmd/start-app"
	"github.com/spf13/cobra"
)

func LoadStartApp(root *cobra.Command) {
	// if runtime.GOOS == constants.RuntimeWindows {
	// 	return
	// }

	root.AddCommand(app.Cmd)
}
