package intercept

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "intercept",
	Short: "intercept app to tunnel trafic to your device",
	Long:  `use this command to intercept an app to tunnel trafic to your device`,
}

func init() {

	fileclient.OnlyInsideBox(startCmd)
	fileclient.OnlyInsideBox(stopCmd)

	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)

	Cmd.Aliases = append(startCmd.Aliases, "inc")
}
