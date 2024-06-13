package auth

import (
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/kloudlite/kl/domain/client"
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout from kloudlite",
	Example: `# Logout from kloudlite
{cmd} auth logout`,
	Run: func(cmd *cobra.Command, args []string) {

		p, err := proxy.NewProxy(false)
		if err != nil {
			fn.PrintError(err)
			return
		}
		p.Stop()

		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.StopAll(); err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.Logout(); err != nil {
			fn.Log(err)
			return
		}

		configFolder, err := client.GetConfigFolder()
		if err != nil {
			fn.Log(err)
			return
		}

		if err := os.RemoveAll(configFolder); err != nil {
			fn.Log(err)
			return
		}

		fn.Log(`successfully logged out.

but the mounted configs, secrets and kl-config stil there, so if you want to also clear this you have clean these manually. 
		`)
	},
}
