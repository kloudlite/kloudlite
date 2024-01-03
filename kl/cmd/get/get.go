package get

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "get",
	Short: "get [ config | secret ] entries",
	Long: `get config/secret entries.

Examples:
  # get config table
  kl get config <configname>

	# get secret table
  kl get secret <secretname>

	# get config/secret in yaml format
  kl get [command] <name> -o yaml

	# get config/secret in json format
  kl get [command] <name> -o json
`,
}

func init() {
	Cmd.AddCommand(configCmd)
	Cmd.AddCommand(secretCmd)
	// Cmd.AddCommand(appsCmd)
}
