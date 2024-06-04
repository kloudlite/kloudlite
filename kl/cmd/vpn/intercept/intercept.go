package intercept

import (
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/flags"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "intercept",
	Short: "intercept app to tunnel trafic to your device",
	Long:  `use this command to intercept an app to tunnel trafic to your device`,
}

func init() {

	if !flags.IsDev() {
		client.OnlyInsideBox(startCmd)
	}

	Cmd.AddCommand(startCmd)

	if !flags.IsDev() {
		client.OnlyInsideBox(stopCmd)
	}

	Cmd.AddCommand(stopCmd)

	Cmd.Aliases = append(startCmd.Aliases, "inc")
}
