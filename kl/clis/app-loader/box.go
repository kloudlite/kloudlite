//go:build box

package apploader

import (
	"runtime"

	"github.com/kloudlite/kl/constants"
	"github.com/spf13/cobra"
)

func LoadStartApp(root *cobra.Command) {
	if runtime.GOOS == constants.RuntimeWindows {
		return
	}

}
