package set_base_url

import (
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "setbaseurl",
	Short:   "set base url for the cli",
	Example: fn.Desc("{cmd} status"),
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {

		if b := functions.ParseBoolFlag(cmd, "reset"); b {
			if err := fileclient.SaveBaseURL(constants.DefaultBaseURL); err != nil {
				fn.PrintError(err)
			} else {
				fn.Log("Base url reset successfully")
			}

			return
		}

		b := functions.ParseBoolFlag(cmd, "check")
		if b {
			fn.Println(constants.BaseURL)
			return
		}

		if len(args) == 0 {
			fn.Log(text.Yellow("Please provide a base url"))
			return
		}

		if err := fileclient.SaveBaseURL(args[0]); err != nil {
			fn.PrintError(err)
		} else {
			fn.Log("Base url set successfully")
		}
	},
}

func init() {
	Cmd.Flags().BoolP("check", "c", false, "check the current base url")
	Cmd.Flags().BoolP("reset", "r", false, "reset the base url to default")
}
