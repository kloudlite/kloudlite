package auth

import (
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout from kloudlite",
	Example: `# Logout from kloudlite
{cmd} auth logout`,
	Run: func(*cobra.Command, []string) {
		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := fc.Logout(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}
