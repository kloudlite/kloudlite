package device

import (
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var setNamespaceCmd = &cobra.Command{
	Use:   "set-namespace",
	Short: "set namespace of selected device",
	Long: `Set namespace of selected device
Examples:
  # set namespace of selected device
	kl device set-namespace <namespace>

	# alternative way
	kl device set-ns <namespace>

  # set namespace of selected device to selected environment
`,
	Run: func(_ *cobra.Command, args []string) {

		ns := ""

		if len(args) > 0 {
			ns = args[0]
		}

		if ns == "" {
			e, err := server.EnsureEnv(nil)
			if err != nil {
				fn.PrintError(err)
				return
			}

			ns = e.TargetNs
		}

		if err := server.UpdateDeviceNS(ns); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("namespace updated successfully")
	},
}

func init() {
	setNamespaceCmd.Aliases = append(setNamespaceCmd.Aliases, "set-ns")
}
